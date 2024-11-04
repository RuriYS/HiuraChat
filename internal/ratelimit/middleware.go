package ratelimit

import (
	"context"
	"time"
)

type RateLimitMiddleware struct {
	limiter *RateLimiter
}

func NewMiddleware(limiter *RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

type RateLimitInfo struct {
	Remaining  int
	Reset      time.Time
	RetryAfter time.Duration
}

func (m *RateLimitMiddleware) Handle(ctx context.Context, route string, handler func() error) error {
	if waitTime, allowed := m.limiter.TryAcquire(route); !allowed {
		// If we're configured to wait, wait for the required time
		if m.limiter.waitForTokens {
			time.Sleep(waitTime)
		} else {
			return &RateLimitError{
				Route:      route,
				RetryAfter: waitTime,
			}
		}
	}

	return handler()
}

type RateLimitError struct {
	Route      string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded for route: " + e.Route + ", retry after: " + e.RetryAfter.String()
}
