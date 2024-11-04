package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"hiurachat/internal/ratelimit"
	"hiurachat/internal/types"
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

	if c.middleware == nil {
		return c.performWrite(v, payload)
	}

	var route string
	if msg, ok := v.(types.Message); ok {
		route = msg.Action
	} else {
		route = "default"
	}

	err = c.middleware.Handle(context.Background(), route, func() error {
		return c.performWrite(v, payload)
	})

	if err != nil {
		if rateLimitErr, ok := err.(*ratelimit.RateLimitError); ok {
			c.logger.Debug("Rate limited on route %s, retry after %v", rateLimitErr.Route, rateLimitErr.RetryAfter)
			time.Sleep(rateLimitErr.RetryAfter)
			return c.WriteJSON(v)
		}
		return err
	}

	return nil
}

func (c *Client) performWrite(v interface{}, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

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
