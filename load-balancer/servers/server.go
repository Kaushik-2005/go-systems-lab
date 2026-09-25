package servers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// Server is one backend that can receive proxied requests.
type Server struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy

	mu      sync.RWMutex
	healthy bool
	active  atomic.Int64
}

func New(address string) (*Server, error) {
	backendURL, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	if backendURL.Scheme == "" || backendURL.Host == "" {
		return nil, &InvalidURL{Value: address}
	}

	server := &Server{
		URL:     backendURL,
		Proxy:   httputil.NewSingleHostReverseProxy(backendURL),
		healthy: true,
	}

	server.Proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, _ error) {
		server.SetHealthy(false)
		http.Error(writer, "backend unavailable", http.StatusBadGateway)
	}

	return server, nil
}

func (s *Server) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthy
}

func (s *Server) SetHealthy(healthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthy = healthy
}

func (s *Server) CheckHealth(client *http.Client) bool {
	healthURL := *s.URL
	healthURL.Path = "/health"

	response, err := client.Get(healthURL.String())
	if err != nil {
		return false
	}
	defer response.Body.Close()

	return response.StatusCode == http.StatusOK
}

func (s *Server) StartRequest() {
	s.active.Add(1)
}

func (s *Server) FinishRequest() {
	s.active.Add(-1)
}

func (s *Server) ActiveRequests() int64 {
	return s.active.Load()
}

type InvalidURL struct {
	Value string
}

func (e *InvalidURL) Error() string {
	return "backend must be an absolute URL: " + e.Value
}
