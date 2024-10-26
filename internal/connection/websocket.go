package connection

import (
	"encoding/json"
	"fmt"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	botID     string
	logger    *logger.Logger
	wsUrl     string
	writeMu   sync.Mutex
	lastWrite time.Time
}

func New(logger *logger.Logger, wsUrl string) (*Client, error) {
	return &Client{
		logger:    logger,
		wsUrl:     wsUrl,
		lastWrite: time.Now().Add(-1 * time.Second),
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
    payload, err := json.Marshal(v)
    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %w", err)
    }

    if c.conn == nil {
        return fmt.Errorf("connection is nil")
    }

    c.writeMu.Lock()
    defer c.writeMu.Unlock()

    since := time.Since(c.lastWrite)
    if since < time.Second {
        time.Sleep(time.Second - since)
    }

    c.logger.Debug("sending payload: %s", string(payload))

    err = c.conn.WriteJSON(v)
    if err != nil {
        c.logger.Error("failed to write JSON: %v", err)
        return fmt.Errorf("failed to write to websocket: %w", err)
    }

    c.lastWrite = time.Now()
    return nil
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
