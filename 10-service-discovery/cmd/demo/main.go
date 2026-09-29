package main

import (
	"context"
	"fmt"
	"time"

	"service-discovery/registry"
)

func main() {
	cleanupContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := registry.New()
	ttl := 2 * time.Second
	discovery.StartCleanup(cleanupContext, 100*time.Millisecond)

	discovery.Register("payments", registry.Instance{
		ID:      "payments-1",
		Address: "localhost:9001",
	}, ttl)

	discovery.Register("payments", registry.Instance{
		ID:      "payments-2",
		Address: "localhost:9002",
	}, ttl)

	instances := discovery.Lookup("payments")

	fmt.Println("payments instances:")
	for _, instance := range instances {
		fmt.Printf("- id=%s address=%s\n", instance.ID, instance.Address)
	}

	time.Sleep(1500 * time.Millisecond)

	heartbeat := discovery.Heartbeat("payments", "payments-1", ttl)
	fmt.Printf("heartbeat payments-1: success=%t\n", heartbeat)

	time.Sleep(1000 * time.Millisecond)

	instances = discovery.Lookup("payments")
	fmt.Printf("payments instances after expiry: %d\n", len(instances))
	for _, instance := range instances {
		fmt.Printf("- id=%s address=%s\n", instance.ID, instance.Address)
	}

	removed := discovery.Deregister("payments", "payments-1")
	fmt.Printf("deregister payments-1: success=%t\n", removed)
}
