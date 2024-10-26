package handler

import (
	"hiurachat/internal/connection"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
	"strings"
)

type MessageHandler struct {
    logger         *logger.Logger
    prefix         string
    responsePrefix string
    conn          *connection.Client
    commands      map[string]types.Command
}

func New(logger *logger.Logger) *MessageHandler {
    return &MessageHandler{
        logger:         logger,
        prefix:         "$",
        responsePrefix: "[BOT: Hermit]",
        commands:       make(map[string]types.Command),
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
            return
        }

        if response.ConnectionId != "" {
            conn.SetBotID(response.ConnectionId)
            h.logger.Info("Connected as: %s (%s)", response.Name, response.ConnectionId)
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
                h.SendMessage(response)
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
