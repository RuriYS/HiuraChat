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
}

func New(logger *logger.Logger) (*Client, error) {
	return &Client{
		logger: logger,
	}, nil
}

func (c *Client) Connect() error {
	url := "wss://gddr51pi43.execute-api.ap-southeast-1.amazonaws.com/dev"
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
