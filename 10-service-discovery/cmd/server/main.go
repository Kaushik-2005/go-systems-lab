package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"service-discovery/registry"
)

func main() {
	discovery := registry.New()

	cleanupContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery.StartCleanup(cleanupContext, 100*time.Millisecond)

	discovery.Register(
		"payments",
		registry.Instance{
			ID:      "payments-1",
			Address: "localhost:9001",
		},
		30*time.Second,
	)

	http.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.Header().Set("Allow", http.MethodPost)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		serviceName := request.URL.Query().Get("service")
		instanceID := request.URL.Query().Get("id")
		address := request.URL.Query().Get("address")

		if serviceName == "" || instanceID == "" || address == "" {
			http.Error(
				writer,
				"missing service, id, or address query parameter",
				http.StatusBadRequest,
			)
			return
		}

		discovery.Register(
			serviceName,
			registry.Instance{
				ID:      instanceID,
				Address: address,
			},
			30*time.Second,
		)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		json.NewEncoder(writer).Encode(map[string]string{
			"status": "registered",
			"id":     instanceID,
		})
	})

	http.HandleFunc("/lookup", func(writer http.ResponseWriter, request *http.Request) {
		serviceName := request.URL.Query().Get("service")

		if serviceName == "" {
			http.Error(writer, "missing service query parameter", http.StatusBadRequest)
			return
		}

		instance := discovery.Lookup(serviceName)

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(instance)
	})

	http.HandleFunc("/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		serviceName := request.URL.Query().Get("service")
		instanceID := request.URL.Query().Get("id")

		if serviceName == "" || instanceID == "" {
			http.Error(
				writer,
				"missing service or id query parameter",
				http.StatusBadRequest,
			)
			return
		}

		renewed := discovery.Heartbeat(
			serviceName,
			instanceID,
			30*time.Second,
		)

		if !renewed {
			http.Error(writer, "instance not found or expired", http.StatusNotFound)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]string{
			"status": "renewed",
		})
	})

	http.HandleFunc("/deregister", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete {
			writer.Header().Set("Allow", http.MethodDelete)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		serviceName := request.URL.Query().Get("service")
		instanceID := request.URL.Query().Get("id")

		if serviceName == "" || instanceID == "" {
			http.Error(
				writer,
				"missing service or id query parameter",
				http.StatusBadRequest,
			)
			return
		}

		removed := discovery.Deregister(serviceName, instanceID)
		if !removed {
			http.Error(writer, "instance not found", http.StatusNotFound)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]string{
			"status": "deregistered",
			"id":     instanceID,
		})
	})

	log.Println("service discovery listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
