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
	conn                 *websocket.Conn // Underlying WebSocket connection
	botID                string          // Unique identifier for the bot
	logger               *logger.Logger  // Logger instance for debugging and error reporting
	wsUrl                string          // WebSocket server URL
	writeMu              sync.Mutex      // Mutex for synchronizing write operations
	lastWrite            time.Time       // Timestamp of the last write operation
	isConnected          bool            // Current connection status
	done                 chan struct{}   // Channel for signaling shutdown
	reconnecting         bool            // Flag indicating if reconnection is in progress
	maxReconnectAttempts int             // Maximum number of reconnection attempts
	readTimeout          time.Duration   // Configurable read timeout
	writeTimeout         time.Duration   // Configurable write timeout
	pingInterval         time.Duration   // Configurable ping interval
	pingTimeout          time.Duration   // Timeout for ping responses
	messageChannel       chan types.Response
}

// New creates a new WebSocket client instance with the specified logger and WebSocket URL.
// Returns an initialized client and any error that occurred during creation.
func New(logger *logger.Logger, wsUrl string) (*Client, error) {
	return &Client{
		logger:               logger,
		wsUrl:                wsUrl,
		lastWrite:            time.Now().Add(-1 * time.Second),
		done:                 make(chan struct{}),
		maxReconnectAttempts: 10,
		readTimeout:          2 * time.Minute,
		writeTimeout:         10 * time.Second,
		pingInterval:         30 * time.Second,
		pingTimeout:          5 * time.Second,
		messageChannel:       make(chan types.Response, 100),
	}, nil
}

// Connect establishes a WebSocket connection to the server.
// It sets up connection timeouts and initializes connection monitoring.
// Returns an error if the connection attempt fails.
func (c *Client) Connect() error {
	url := c.wsUrl

	// Configure connection timeout
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	dialer.EnableCompression = true

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}

	// Set connection deadlines
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	c.conn = conn
	c.isConnected = true

	// Start connection monitor
	go c.monitorConnection()

	return nil
}

// monitorConnection continuously monitors the WebSocket connection for incoming messages and status
func (c *Client) monitorConnection() {
	c.conn.SetPingHandler(func(appData string) error {
		c.logger.Debug("Received ping")
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		return c.conn.WriteControl(
			websocket.PongMessage,
			[]byte{},
			time.Now().Add(c.pingTimeout),
		)
	})

	c.conn.SetPongHandler(func(appData string) error {
		c.logger.Debug("Received pong")
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		return nil
	})

	for {
		select {
		case <-c.done:
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))

			var response types.Response
			err := c.conn.ReadJSON(&response)

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("Unexpected close error: %v", err)
				} else {
					c.logger.Debug("Connection closed: %v", err)
				}
				c.handleDisconnect()
				return
			}

			// Log all received messages for debugging
			if data, err := json.Marshal(response); err == nil {
				c.logger.Debug("Received message: %s", string(data))
			}

			// Send the message to the channel for processing
			select {
			case c.messageChannel <- response:
			default:
				c.logger.Warn("Message channel full, dropping message")
			}
		}
	}
}

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

// handleDisconnect manages the disconnection process and initiates reconnection if appropriate.
// It ensures thread-safe state updates and prevents multiple simultaneous reconnection attempts.
func (c *Client) handleDisconnect() {
	c.writeMu.Lock()
	wasConnected := c.isConnected
	c.isConnected = false
	c.writeMu.Unlock()

	if wasConnected && !c.reconnecting {
		go c.reconnectWithBackoff()
	}
}

// reconnectWithBackoff implements an exponential backoff strategy for reconnection attempts.
// It will try to reconnect multiple times with increasing delays between attempts.
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

	for attempt := 0; attempt < c.maxReconnectAttempts && !c.isConnected; attempt++ {
		select {
		case <-c.done:
			return
		default:
			delay := backoff[min(attempt, len(backoff)-1)]
			c.logger.Debug("Attempting reconnection in %v (attempt %d/%d)",
				delay, attempt+1, c.maxReconnectAttempts)

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

// StartHeartbeat initiates a heartbeat goroutine that sends periodic ping messages
// to keep the connection alive. The interval parameter determines the frequency of heartbeats.
func (c *Client) StartHeartbeat(interval time.Duration) {
	if interval < c.pingInterval {
		interval = c.pingInterval
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if !c.isConnected {
					continue
				}

				c.logger.Debug("Sending heartbeat")
				deadline := time.Now().Add(c.writeTimeout)
				if err := c.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
					c.logger.Error("Heartbeat failed: %v", err)
					c.handleDisconnect()
				} else {
					c.logger.Debug("Heartbeat sent")
					// Reset read deadline after successful heartbeat
					c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
				}

			case <-c.done:
				ticker.Stop()
				return
			}
		}
	}()
}

// WriteControl sends a WebSocket control message with the specified parameters.
// It ensures thread-safe access to the connection for control message writing.
func (c *Client) WriteControl(messageType int, data []byte, deadline time.Time) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteControl(messageType, data, deadline)
}

// reconnect attempts to reestablish the WebSocket connection after a disconnection.
// It closes the existing connection if necessary and attempts to create a new one.
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

// Close gracefully shuts down the WebSocket connection and stops all goroutines.
// It sends a close frame to the server and cleans up resources.
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

// RequestID sends a request to the server to get the bot's ID.
// This is typically used during the initial connection or reconnection.
func (c *Client) RequestID() error {
	msg := types.Message{
		Action: "getId",
	}
	return c.WriteJSON(msg)
}

// WriteJSON sends a JSON message to the server with rate limiting and error handling.
// It ensures thread-safe access to the connection and implements backoff between writes.
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

	// Rate limiting
	since := time.Since(c.lastWrite)
	if since < time.Second {
		time.Sleep(time.Second - since)
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	c.logger.Debug("Payload: %s", string(payload))

	if err := c.conn.WriteJSON(v); err != nil {
		c.logger.Error("failed to write JSON: %v", err)
		c.handleDisconnect()
		return fmt.Errorf("failed to write to websocket: %w", err)
	}

	c.lastWrite = time.Now()
	return nil
}

// min returns the minimum of two integers.
// Used for calculating backoff delays.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ReadJSON reads and decodes a JSON message from the WebSocket connection.
// Returns an error if the connection is nil or if reading/decoding fails.
func (c *Client) ReadJSON(v interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	return c.conn.ReadJSON(v)
}

// GetBotID returns the current bot ID.
func (c *Client) GetBotID() string {
	return c.botID
}

// SetBotID sets the bot ID to the specified value.
func (c *Client) SetBotID(id string) {
	c.botID = id
}
