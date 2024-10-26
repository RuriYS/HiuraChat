package connection

import (
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	botID  string
	mu     sync.Mutex
	logger *logger.Logger
	wsUrl  string
}

func New(logger *logger.Logger, wsUrl string) (*Client, error) {
	return &Client{
		logger: logger,
		wsUrl:  wsUrl,
	}, nil
}

func (c *Client) Connect() error {
	url := c.wsUrl
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
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

func (c *Client) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *Client) ReadJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.ReadJSON(v)
}

func (c *Client) GetBotID() string {
	return c.botID
}

func (c *Client) SetBotID(id string) {
	c.botID = id
}
