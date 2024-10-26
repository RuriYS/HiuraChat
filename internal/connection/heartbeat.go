package connection

import (
	"time"
)

func (c *Client) StartHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.logger.Debug("Sending heartbeat")

			if err := c.RequestID(); err != nil {
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

	if c.conn != nil {
		c.conn.Close()
	}

	err := c.Connect()
	if err != nil {
		return err
	}

	return c.RequestID()
}
