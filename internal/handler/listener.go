package handler

import (
	"encoding/json"
	"fmt"
	"hiurachat/internal/connection"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
	"strings"
	"time"
)

type MessageHandler struct {
	logger         *logger.Logger
	prefix         string
	responsePrefix string
	conn           *connection.Client
	commands       map[string]types.Command
	bot            types.Latency
}

func New(logger *logger.Logger, prefix string, rprefix string, bot types.Latency) *MessageHandler {
	return &MessageHandler{
		logger:         logger,
		prefix:         prefix,
		responsePrefix: rprefix,
		commands:       make(map[string]types.Command),
		bot:            bot,
	}
}

func (h *MessageHandler) SetCommands(commands map[string]types.Command) {
	h.commands = commands
}

func (h *MessageHandler) GetPrefix() string {
	return h.prefix
}

func (h *MessageHandler) GetResponsePrefix() string {
	return h.responsePrefix
}

func (h *MessageHandler) HandleCommand(commandStr string, args []string) (string, bool) {
	commandName := strings.TrimPrefix(commandStr, h.prefix)

	command, exists := h.commands[commandName]
	if !exists {
		return "", false
	}

	return command.Execute(args)
}

func (h *MessageHandler) Listen(conn *connection.Client) {
	h.conn = conn
	for {
		var response types.Response
		err := conn.ReadJSON(&response)

		if err != nil {
			h.logger.Error("Error reading: %v", err)
			time.Sleep(time.Second)
			continue
		}

		data, err := json.Marshal(response)
		{
			if err != nil {
				h.logger.Debug(err.Error())
			}
			h.logger.Debug("Event: " + string(data))
		}

		if response.ConnectionId != "" {
			if conn.GetBotID() == "" {
				conn.SetBotID(response.ConnectionId)
				h.logger.Info("Connected as: %s (%s)", response.Name, conn.GetBotID())
			}

			pingTime := h.bot.GetLatency()
			if !pingTime.IsZero() {
				latency := time.Since(pingTime)
				h.bot.SetLatency(time.Time{})
				err := h.SendMessage(fmt.Sprintf("%s Pong! (Latency: %.2fms)", h.GetResponsePrefix(), float64(latency.Microseconds())/1000.0))
				if err != nil {
					h.logger.Error("Failed to send ping response: %s", err)
				}
			}
			continue
		}

		if response.Message == "" || response.Sender == conn.GetBotID() {
			continue
		}

		parts := strings.Fields(response.Message)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		if strings.HasPrefix(command, h.prefix) {
			if response, ok := h.HandleCommand(command, args); ok {
				if err := h.SendMessage(response); err != nil {
					h.logger.Error("Failed to send message: %v", err)
				}
			}
		}

		h.logger.Info("%s: %s", response.SenderName, response.Message)
	}
}

func (h *MessageHandler) SendMessage(message string) error {
	msg := types.Message{
		Action: "sendMessage",
		Data: &types.MessageData{
			Message: message,
		},
	}
	return h.conn.WriteJSON(msg)
}
