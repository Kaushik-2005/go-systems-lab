package main

import (
	"flag"
	"fmt"
	"time"

	"rate-limiter/algorithms"
	"rate-limiter/ratelimiter"
)

func main() {
	limit := flag.Int("limit", 3, "maximum requests per window")
	window := flag.Duration("window", 5*time.Second, "window duration")
	algorithm := flag.String("algorithm", "fixed", "algorithm: fixed, sliding, or sliding-counter")
	capacity := flag.Int("capacity", 3, "token bucket capacity")
	refillRate := flag.Float64("refill-rate", 1, "token bucket refill rate per second")
	key := flag.String("key", "client-1", "key to rate limit")
	requests := flag.Int("requests", 5, "requests per batch")
	batches := flag.Int("batches", 1, "number of batches to send")
	wait := flag.Duration("wait", 0, "pause between batches")
	flag.Parse()

	var limiter ratelimiter.Limiter
	switch *algorithm {
	case "fixed":
		limiter = algorithms.NewFixedWindow(*limit, *window)
	case "sliding":
		limiter = algorithms.NewSlidingWindowLog(*limit, *window)
	case "sliding-counter":
		limiter = algorithms.NewSlidingWindowCounter(*limit, *window)
	case "token":
		limiter = algorithms.NewTokenBucket(*capacity, *refillRate)
	default:
		panic("algorithm must be fixed, sliding, sliding-counter, or token")
	}
	requestNumber := 0
	for batch := 1; batch <= *batches; batch++ {
		for request := 1; request <= *requests; request++ {
			requestNumber++
			fmt.Printf("time=%s batch=%d request=%d key=%s allowed=%t\n", time.Now().Format("15:04:05.000"), batch, requestNumber, *key, limiter.Allow(*key))
		}

		if batch < *batches {
			time.Sleep(*wait)
		}
	}
}
