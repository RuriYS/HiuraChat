package connection

import (
	"time"
)

func (c *Client) handleDisconnect() {
	c.writeMu.Lock()
	wasConnected := c.isConnected
	c.isConnected = false
	c.writeMu.Unlock()

	if wasConnected && !c.reconnecting {
		go c.reconnectWithBackoff()
	}
}

func (c *Client) reconnectWithBackoff() {
	c.reconnecting = true
	defer func() { c.reconnecting = false }()

	backoff := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for attempt := 0; attempt < c.config.MaxReconnectAttempts && !c.isConnected; attempt++ {
		select {
		case <-c.done:
			return
		default:
			delay := backoff[min(attempt, len(backoff)-1)]
			c.logger.Debug("Attempting reconnection in %v (attempt %d/%d)",
				delay, attempt+1, c.config.MaxReconnectAttempts)

			time.Sleep(delay)

			if err := c.reconnect(); err != nil {
				c.logger.Error("Reconnection attempt failed: %v", err)
				continue
			}

			c.logger.Debug("Successfully reconnected")
			return
		}
	}

	if !c.isConnected {
		c.logger.Error("Max reconnection attempts reached, giving up")
	}
}

func (c *Client) reconnect() error {
	if c.conn != nil {
		c.conn.Close()
	}

	if err := c.Connect(); err != nil {
		return err
	}

	return c.RequestID()
}
