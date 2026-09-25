package algorithms

import (
	"sync"
	"time"
)

type bucketState struct {
	tokens   float64
	refilled time.Time
}

// TokenBucket allows bursts up to capacity and refills tokens over time.
type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	refillRate float64
	buckets    map[string]bucketState
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   float64(capacity),
		refillRate: refillRate,
		buckets:    make(map[string]bucketState),
	}
}

func (t *TokenBucket) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	bucket, exists := t.buckets[key]
	if !exists {
		bucket = bucketState{tokens: t.capacity, refilled: now}
	}

	elapsed := now.Sub(bucket.refilled).Seconds()
	bucket.tokens = min(t.capacity, bucket.tokens+elapsed*t.refillRate)
	bucket.refilled = now

	if bucket.tokens < 1 {
		t.buckets[key] = bucket
		return false
	}

	bucket.tokens--
	t.buckets[key] = bucket
	return true
}

func min(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}
