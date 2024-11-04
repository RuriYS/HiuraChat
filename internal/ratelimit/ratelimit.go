package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu            sync.RWMutex
	globalBucket  *TokenBucket
	routeBuckets  map[string]*TokenBucket
	defaultLimit  Rate
	defaultBurst  int
	routeLimits   map[string]Rate
	waitForTokens bool
}

type Rate struct {
	Limit  float64       // Tokens per second
	Burst  int           // Maximum burst size
	Window time.Duration // Time window for rate limiting
}

type TokenBucket struct {
	mu        sync.Mutex
	rate      Rate
	tokens    float64
	lastTime  time.Time
	maxTokens float64
}

func NewRateLimiter(defaultRate Rate, waitForTokens bool) *RateLimiter {
	rl := &RateLimiter{
		globalBucket:  newTokenBucket(defaultRate),
		routeBuckets:  make(map[string]*TokenBucket),
		defaultLimit:  defaultRate,
		routeLimits:   make(map[string]Rate),
		waitForTokens: waitForTokens,
	}
	return rl
}

func newTokenBucket(rate Rate) *TokenBucket {
	return &TokenBucket{
		rate:      rate,
		tokens:    float64(rate.Burst),
		lastTime:  time.Now(),
		maxTokens: float64(rate.Burst),
	}
}

func (tb *TokenBucket) update(now time.Time) {
	elapsed := now.Sub(tb.lastTime)
	tokensPerSecond := tb.rate.Limit / tb.rate.Window.Seconds()
	tb.tokens += elapsed.Seconds() * tokensPerSecond
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastTime = now
}

func (tb *TokenBucket) tryAcquire(wait bool) (time.Duration, bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.update(now)

	if tb.tokens >= 1.0 {
		tb.tokens--
		return 0, true
	}

	tokensPerSecond := tb.rate.Limit / tb.rate.Window.Seconds()
	timeToNext := time.Duration((1.0 - tb.tokens) / tokensPerSecond * float64(time.Second))

	if !wait {
		return timeToNext, false
	}

	time.Sleep(timeToNext)
	tb.tokens = 0
	return timeToNext, true
}

func (rl *RateLimiter) SetRouteLimit(route string, rate Rate) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.routeLimits[route] = rate
	rl.routeBuckets[route] = newTokenBucket(rate)
}

func (rl *RateLimiter) Wait(route string) time.Duration {
	bucket := rl.getBucket(route)
	waitTime, _ := bucket.tryAcquire(true)
	return waitTime
}

func (rl *RateLimiter) TryAcquire(route string) (time.Duration, bool) {
	bucket := rl.getBucket(route)
	return bucket.tryAcquire(false)
}

func (rl *RateLimiter) getBucket(route string) *TokenBucket {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if bucket, exists := rl.routeBuckets[route]; exists {
		return bucket
	}
	return rl.globalBucket
}
