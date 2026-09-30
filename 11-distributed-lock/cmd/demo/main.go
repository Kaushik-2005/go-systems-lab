package main

import (
	"fmt"
	"time"

	"distributed-lock/lock"
)

func main() {
	mutex := lock.New()

	tokenA, acquired := mutex.Acquire("client-a", 500*time.Millisecond)
	fmt.Printf("client-a acquire: success=%t token=%d\n", acquired, tokenA)

	time.Sleep(700 * time.Millisecond)

	tokenB, acquired := mutex.Acquire("client-b", time.Second)
	fmt.Printf("client-b acquire after expiry: success=%t token=%d\n", acquired, tokenB)

	staleRelease := mutex.Release("client-a", tokenA)
	fmt.Printf("client-a stale release: success=%t\n", staleRelease)

	renewed := mutex.Renew("client-b", tokenB, time.Second)
	fmt.Printf("client-b renew: success=%t\n", renewed)

	released := mutex.Release("client-b", tokenB)
	fmt.Printf("client-b release: success=%t\n", released)
}
