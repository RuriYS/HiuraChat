// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot       BotConfig       `yaml:"bot"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	Logger    LoggerConfig    `yaml:"logger"`
}

type BotConfig struct {
	Prefix         string `yaml:"prefix"`
	ResponsePrefix string `yaml:"response_prefix"`
}

type WebSocketConfig struct {
	URL                  string `yaml:"url"`
	WSUrl                string
	MaxReconnectAttempts int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	PingInterval         time.Duration
	PingTimeout          time.Duration
	HandshakeTimeout     time.Duration
	MessageBufferSize    int
}

func DefaultConfig() WebSocketConfig {
	return WebSocketConfig{
		MaxReconnectAttempts: 10,
		ReadTimeout:          2 * time.Minute,
		WriteTimeout:         10 * time.Second,
		PingInterval:         30 * time.Second,
		PingTimeout:          5 * time.Second,
		HandshakeTimeout:     10 * time.Second,
		MessageBufferSize:    100,
	}
}

type LoggerConfig struct {
	Level     string `yaml:"level"`
	UseColors bool   `yaml:"use_colors"`
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
