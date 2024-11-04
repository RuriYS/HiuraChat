package connection

import (
	"fmt"
	"sync"
	"time"

	"hiurachat/internal/config"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"

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
}

func New(logger *logger.Logger, wsUrl string, cfg *config.WebSocketConfig) (*Client, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if wsUrl == "" {
		return nil, fmt.Errorf("websocket URL is required")
	}

	if cfg == nil {
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg
	}
	cfg.WSUrl = wsUrl

	return &Client{
		logger:         logger,
		config:         *cfg,
		lastWrite:      time.Now().Add(-1 * time.Second),
		done:           make(chan struct{}),
		messageChannel: make(chan types.Response, cfg.MessageBufferSize),
	}, nil
}
