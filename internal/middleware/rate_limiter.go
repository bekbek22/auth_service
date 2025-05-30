package middleware

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]int64 // email -> timestamps
	limit    int
	window   int64 // sec
}

func NewRateLimiter(limit int, windowSeconds int64) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]int64),
		limit:    limit,
		window:   windowSeconds,
	}
}

func (r *RateLimiter) Allow(email string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	windowStart := now - r.window

	times := r.attempts[email]

	// Filter only attempts that are still in the window.
	var recent []int64
	for _, t := range times {
		if t >= windowStart {
			recent = append(recent, t)
		}
	}

	if len(recent) >= r.limit {
		return false
	}

	// Allow and record time
	recent = append(recent, now)
	r.attempts[email] = recent
	return true
}
