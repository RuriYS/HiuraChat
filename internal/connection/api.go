package connection

import (
	"fmt"
	"hiurachat/internal/types"

	"github.com/gorilla/websocket"
)

func (c *Client) Listen(handler func(types.Response)) {
	for {
		select {
		case <-c.done:
			return
		case msg := <-c.messageChannel:
			handler(msg)
		}
	}
}

func (c *Client) Close() error {
	close(c.done)
	if c.conn != nil {
		err := c.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			c.logger.Error("Error sending close frame: %v", err)
		}
		return c.conn.Close()
	}
	return nil
}

func (c *Client) RequestID() error {
	msg := types.Message{
		Action: "getId",
	}
	return c.WriteJSON(msg)
}

func (c *Client) ReadJSON(v interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	return c.conn.ReadJSON(v)
}

func (c *Client) GetBotID() string {
	return c.botID
}

func (c *Client) SetBotID(id string) {
	c.botID = id
}
