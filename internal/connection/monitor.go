package connection

import (
	"encoding/json"
	"hiurachat/internal/types"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) monitorConnection() {
	c.setupPingPong()

	for {
		select {
		case <-c.done:
			return
		default:
			if err := c.readMessage(); err != nil {
				c.handleDisconnect()
				return
			}
		}
	}
}

func (c *Client) setupPingPong() {
	c.conn.SetPingHandler(func(appData string) error {
		c.logger.Debug("Received ping")
		c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
		return c.conn.WriteControl(
			websocket.PongMessage,
			[]byte{},
			time.Now().Add(c.config.PingTimeout),
		)
	})

	c.conn.SetPongHandler(func(appData string) error {
		c.logger.Debug("Received pong")
		c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
		return nil
	})
}

func (c *Client) readMessage() error {
	c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))

	var response types.Response
	if err := c.conn.ReadJSON(&response); err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			c.logger.Error("Unexpected close error: %v", err)
		} else {
			c.logger.Debug("Connection closed: %v", err)
		}
		return err
	}

	if data, err := json.Marshal(response); err == nil {
		c.logger.Debug("Received message: %s", string(data))
	}

	select {
	case c.messageChannel <- response:
	default:
		c.logger.Warn("Message channel full, dropping message")
	}

	return nil
}
