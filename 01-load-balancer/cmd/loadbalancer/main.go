package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"load-balancer/algorithms"
	"load-balancer/loadbalancer"
	"load-balancer/servers"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "address where the load balancer listens")
	backendList := flag.String("backends", "http://localhost:9000", "comma-separated backend URLs")
	algorithmName := flag.String("algorithm", "roundrobin", "selection algorithm: roundrobin or random")
	requestTimeout := flag.Duration("request-timeout", 5*time.Second, "maximum time allowed for one forwarded request")
	flag.Parse()

	backends := make([]*servers.Server, 0)
	for _, address := range strings.Split(*backendList, ",") {
		backend, err := servers.New(strings.TrimSpace(address))
		if err != nil {
			log.Fatal(err)
		}
		backends = append(backends, backend)
	}

	selector, err := selectorFor(*algorithmName)
	if err != nil {
		log.Fatal(err)
	}

	balancer := loadbalancer.New(backends, selector, *requestTimeout)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	balancer.StartHealthChecks(ctx, 2*time.Second)

	log.Printf("load balancer listening on %s with %d backends", *listenAddr, len(backends))

	server := &http.Server{
		Addr:    *listenAddr,
		Handler: balancer,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("load balancer failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func selectorFor(name string) (algorithms.Selector, error) {
	switch name {
	case "roundrobin":
		return &algorithms.RoundRobin{}, nil
	case "random":
		return algorithms.NewRandom(), nil
	case "leastconnections":
		return &algorithms.LeastConnections{}, nil
	default:
		return nil, fmt.Errorf("unknown algorithm %q", name)
	}
}
