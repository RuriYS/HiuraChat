package connection

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) Connect() error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.config.HandshakeTimeout
	dialer.EnableCompression = true

	conn, _, err := dialer.Dial(c.config.WSUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))

	c.conn = conn
	c.isConnected = true
	go c.monitorConnection()

	return nil
}
