package main

import (
	"fmt"
	"time"

	"distributed-cache/cache"
)

func main() {
	node, err := cache.NewNode(2)
	if err != nil {
		panic(err)
	}

	node.Set("language", "Go", 0)
	node.Set("database", "systems", 0)
	value, found := node.Get("language")
	fmt.Printf("get language: value=%q found=%t\n", value, found)

	node.Set("region", "asia", 0)
	_, databaseFound := node.Get("database")
	fmt.Printf("after capacity eviction: database_found=%t\n", databaseFound)

	node.Set("temporary", "value", 100*time.Millisecond)
	_, beforeExpiry := node.Get("temporary")
	time.Sleep(150 * time.Millisecond)
	_, afterExpiry := node.Get("temporary")
	fmt.Printf("ttl: before_expiry=%t after_expiry=%t\n", beforeExpiry, afterExpiry)

	cluster, err := cache.NewCluster([]string{"cache-a", "cache-b", "cache-c"}, 10, 20, 2)
	if err != nil {
		panic(err)
	}

	keys := []string{"user-1", "user-2", "user-3", "order-1", "order-2", "profile-1"}
	before := routeAll(cluster, keys)
	fmt.Printf("cluster nodes: %v\n", cluster.NodeNames())

	replicas, err := cluster.ReplicaNodes("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("user-1 replicas: %v\n", replicas)
	fmt.Printf("before adding cache-d: %v\n", before)

	if _, err := cluster.Set("user-1", "Kaushik", 0); err != nil {
		panic(err)
	}
	value, found, owner, err := cluster.Get("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("cluster get: value=%q found=%t node=%s\n", value, found, owner)

	if err := cluster.SetNodeHealth("cache-b", false); err != nil {
		panic(err)
	}
	value, found, owner, err = cluster.Get("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("after primary failure: value=%q found=%t node=%s\n", value, found, owner)

	if err := cluster.SetNodeHealth("cache-b", true); err != nil {
		panic(err)
	}
	if err := cluster.AddNode("cache-d"); err != nil {
		panic(err)
	}
	after := routeAll(cluster, keys)
	migrated, err := cluster.Rebalance()
	if err != nil {
		panic(err)
	}
	fmt.Printf("after adding cache-d: moved=%d/%d records_migrated=%d\n", countMoved(before, after), len(keys), migrated)

	if err := cluster.RemoveNode("cache-d"); err != nil {
		panic(err)
	}
	value, found, owner, err = cluster.Get("user-1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("after removing cache-d: value=%q found=%t node=%s\n", value, found, owner)
}

func routeAll(cluster *cache.Cluster, keys []string) map[string]string {
	routes := make(map[string]string, len(keys))
	for _, key := range keys {
		node, err := cluster.NodeFor(key)
		if err != nil {
			panic(err)
		}
		routes[key] = node
	}
	return routes
}

func countMoved(before, after map[string]string) int {
	moved := 0
	for key, node := range before {
		if after[key] != node {
			moved++
		}
	}
	return moved
}
