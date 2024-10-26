package connection

import (
	"time"
)

const HeartbeatInterval = 30 * time.Second

func (c *Client) StartHeartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	go func() {
		for range ticker.C {
			err := c.RequestID()
			if err != nil {
				c.logger.Error("Heartbeat failed: %v", err)
				if err := c.reconnect(); err != nil {
					c.logger.Error("Failed to reconnect: %v", err)
					continue
				}
			}
			c.logger.Debug("Heartbeat sent")
		}
	}()
}

func (c *Client) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	err := c.Connect()
	if err != nil {
		return err
	}

	return c.RequestID()
}
