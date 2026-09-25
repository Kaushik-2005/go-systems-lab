package ratelimiter

// Limiter decides whether an operation for a key may proceed.
type Limiter interface {
	Allow(key string) bool
}
