package config

import (
	"fmt"
	"hiurachat/internal/ratelimit"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot struct {
		Prefix         string `yaml:"prefix"`
		ResponsePrefix string `yaml:"response_prefix"`
	} `yaml:"bot"`

	WebSocket struct {
		URL string `yaml:"url"`
	} `yaml:"websocket"`

	Logger struct {
		Level     string `yaml:"level"`
		UseColors bool   `yaml:"use_colors"`
	} `yaml:"logger"`
}

type WebSocketConfig struct {
	URL                  string
	MaxReconnectAttempts int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	PingInterval         time.Duration
	PingTimeout          time.Duration
	HandshakeTimeout     time.Duration
	MessageBufferSize    int
	RateLimit            struct {
		Enabled       bool
		Global        GlobalRateLimit
		WaitForTokens bool
		RouteLimits   map[string]ratelimit.Rate
	}
}

type GlobalRateLimit struct {
	Messages int
	Window   time.Duration
	Burst    int
}

func LoadConfig(path string) (*Config, error) {
	if !filepath.IsAbs(path) {
		workDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %v", err)
		}
		path = filepath.Join(workDir, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func (c *Config) GetWebSocketConfig() *WebSocketConfig {
	cfg := &WebSocketConfig{
		URL:                  c.WebSocket.URL,
		MaxReconnectAttempts: 10,
		ReadTimeout:          2 * time.Minute,
		WriteTimeout:         10 * time.Second,
		PingInterval:         30 * time.Second,
		PingTimeout:          5 * time.Second,
		HandshakeTimeout:     10 * time.Second,
		MessageBufferSize:    100,
	}

	cfg.RateLimit.Enabled = true
	cfg.RateLimit.Global = GlobalRateLimit{
		Messages: 15,
		Window:   30 * time.Second,
		Burst:    3,
	}

	cfg.RateLimit.WaitForTokens = true

	cfg.RateLimit.RouteLimits = map[string]ratelimit.Rate{
		"default": {
			Limit:  float64(cfg.RateLimit.Global.Messages),
			Burst:  cfg.RateLimit.Global.Burst,
			Window: cfg.RateLimit.Global.Window,
		},
		"getId": {
			Limit:  1,
			Burst:  1,
			Window: 5 * time.Second,
		},
	}

	return cfg
}
