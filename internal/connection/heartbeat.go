package connection

import (
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) StartHeartbeat(interval time.Duration) {
	if interval < c.config.PingInterval {
		interval = c.config.PingInterval
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
				deadline := time.Now().Add(c.config.WriteTimeout)
				if err := c.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
					c.logger.Error("Heartbeat failed: %v", err)
					c.handleDisconnect()
				} else {
					c.logger.Debug("Heartbeat sent")
					c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
				}

			case <-c.done:
				ticker.Stop()
				return
			}
		}
	}()
}
