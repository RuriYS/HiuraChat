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
		GlobalRate    float64
		GlobalBurst   int
		WaitForTokens bool
		RouteLimits   map[string]ratelimit.Rate
	}
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
	cfg.RateLimit.GlobalRate = 100
	cfg.RateLimit.GlobalBurst = 5
	cfg.RateLimit.WaitForTokens = true
	cfg.RateLimit.RouteLimits = map[string]ratelimit.Rate{
		"sendMessage": {
			Limit:  2,
			Burst:  1,
			Window: time.Second,
		},
		"getId": {
			Limit:  0.2, // 1 request per 5 seconds
			Burst:  1,
			Window: 5 * time.Second,
		},
	}

	return cfg
}
