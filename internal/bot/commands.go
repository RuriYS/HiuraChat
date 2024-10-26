package bot

import (
	"fmt"
	"hiurachat/internal/types"
	"strings"
	"time"
)

func (b *Bot) initializeCommands() {
	b.commands = map[string]types.Command{
		"ping": {
			Name:        "ping",
			Description: "Check bot latency",
			Execute: func(args []string) (string, bool) {
				b.pingTime = time.Now()
				if err := b.conn.RequestID(); err != nil {
					return fmt.Sprintf("%s Failed to ping", b.handler.GetResponsePrefix()), true
				}
				return "", false
			},
		},
		"echo": {
			Name:        "echo",
			Description: "Echo back your message",
			Execute: func(args []string) (string, bool) {
				if len(args) >= 1 {
					return fmt.Sprintf("%s %s", b.handler.GetResponsePrefix(), strings.Join(args, " ")), true
				}
				return "", false
			},
		},
		"help": {
			Name:        "Help",
			Description: "Help",
			Execute: func(args []string) (string, bool) {
				return fmt.Sprintf("%s - %s", b.commands[args[0]].Name, b.commands[args[0]].Description), true
			},
		},
	}
}

func (b *Bot) HandleCommand(commandStr string, args []string) (string, bool) {
	commandName := strings.TrimPrefix(commandStr, b.handler.GetPrefix())

	command, exists := b.commands[commandName]
	if !exists {
		return "", false
	}

	return command.Execute(args)
}
