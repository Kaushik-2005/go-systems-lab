package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"lru-cache/cache"
)

func main() {
	values := cache.New(2)
	values.Set("language", "Go")
	printState("after Set(language)", values)
	values.Set("database", "PostgreSQL")
	printState("after Set(database)", values)

	values.Get("language")
	printState("after Get(language)", values)
	values.Set("cache", "LRU")
	printState("after Set(cache)", values)
	_, databaseFound := values.Get("database")
	value, languageFound := values.Get("language")
	fmt.Printf("eviction check: language=%q found=%t database_found=%t\n", value, languageFound, databaseFound)

	deleted := values.Delete("language")
	printState("after Delete(language)", values)
	fmt.Printf("delete check: deleted=%t\n", deleted)

	expiring := cache.New(2)
	expiring.SetWithTTL("session", "active", 500*time.Millisecond)
	_, foundBeforeExpiry := expiring.Get("session")
	time.Sleep(600 * time.Millisecond)
	_, foundAfterExpiry := expiring.Get("session")
	fmt.Printf("ttl: found_before_expiry=%t found_after_expiry=%t\n", foundBeforeExpiry, foundAfterExpiry)

	concurrent := cache.New(100)
	var waitGroup sync.WaitGroup
	var misses atomic.Int64
	const workers = 10
	const operationsPerWorker = 100
	for worker := 0; worker < workers; worker++ {
		waitGroup.Add(1)
		go func(worker int) {
			defer waitGroup.Done()
			for request := 0; request < operationsPerWorker; request++ {
				key := fmt.Sprintf("worker-%d", worker)
				concurrent.Set(key, "value")
				if _, ok := concurrent.Get(key); !ok {
					misses.Add(1)
				}
			}
		}(worker)
	}
	waitGroup.Wait()
	fmt.Printf("concurrent access: operations=%d misses=%d entries=%d\n", workers*operationsPerWorker, misses.Load(), concurrent.Len())

	cleanupContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleaned := cache.New(2)
	cleaned.StartCleanup(cleanupContext, 50*time.Millisecond)
	cleaned.SetWithTTL("temporary", "value", 100*time.Millisecond)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("background cleanup: entries=%d\n", cleaned.Len())
}

func printState(label string, values *cache.Cache) {
	fmt.Printf("%-24s order(MRU->LRU)=%v size=%d\n", label, values.Keys(), values.Len())
}
