package main

import (
	"time"
)

const HeartbeatInterval = time.Minute

func startHeartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	go func() {
		logger := getLogger()
		for range ticker.C {
			err := requestId()
			if err != nil {
				logger.Error("Heartbeat failed: %v", err)
				if err := reconnect(); err != nil {
					logger.Error("Failed to reconnect: %v", err)
					continue
				}
			}
			logger.Debug("Heartbeat sent")
		}
	}()
}

func reconnect() error {
	connMutex.Lock()
	defer connMutex.Unlock()

	if conn != nil {
		conn.Close()
	}

	err := connect()
	if err != nil {
		return err
	}

	return requestId()
}