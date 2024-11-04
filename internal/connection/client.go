package connection

import (
	"fmt"
	"hiurachat/internal/config"
	"hiurachat/internal/logger"
	"hiurachat/internal/ratelimit"
	"hiurachat/internal/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn           *websocket.Conn
	botID          string
	logger         *logger.Logger
	config         config.WebSocketConfig
	writeMu        sync.Mutex
	lastWrite      time.Time
	isConnected    bool
	done           chan struct{}
	reconnecting   bool
	messageChannel chan types.Response
	rateLimiter    *ratelimit.RateLimiter
	middleware     *ratelimit.RateLimitMiddleware
}

func New(logger *logger.Logger, wsUrl string, cfg *config.WebSocketConfig) (*Client, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if wsUrl == "" {
		return nil, fmt.Errorf("websocket URL is required")
	}

	var rateLimiter *ratelimit.RateLimiter
	var middleware *ratelimit.RateLimitMiddleware

	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.NewRateLimiter(ratelimit.Rate{
			Limit:  cfg.RateLimit.GlobalRate,
			Burst:  cfg.RateLimit.GlobalBurst,
			Window: time.Second,
		}, cfg.RateLimit.WaitForTokens)

		for route, rate := range cfg.RateLimit.RouteLimits {
			rateLimiter.SetRouteLimit(route, rate)
		}

		middleware = ratelimit.NewMiddleware(rateLimiter)
	}

	client := &Client{
		logger:         logger,
		config:         *cfg,
		lastWrite:      time.Now().Add(-1 * time.Second),
		done:           make(chan struct{}),
		messageChannel: make(chan types.Response, cfg.MessageBufferSize),
		rateLimiter:    rateLimiter,
		middleware:     middleware,
	}

	return client, nil
}
