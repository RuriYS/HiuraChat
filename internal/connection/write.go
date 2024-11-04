package connection

import (
	"encoding/json"
	"fmt"
	"time"
)

func (c *Client) WriteJSON(v interface{}) error {
	if !c.isConnected {
		return fmt.Errorf("not connected")
	}

	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	since := time.Since(c.lastWrite)
	if since < time.Second {
		time.Sleep(time.Second - since)
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	c.logger.Debug("Payload: %s", string(payload))

	if err := c.conn.WriteJSON(v); err != nil {
		c.logger.Error("failed to write JSON: %v", err)
		c.handleDisconnect()
		return fmt.Errorf("failed to write to websocket: %w", err)
	}

	c.lastWrite = time.Now()
	return nil
}

func (c *Client) WriteControl(messageType int, data []byte, deadline time.Time) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteControl(messageType, data, deadline)
}
