package main

import (
	"fmt"
	"strings"
	"time"
)

var commands = make(map[string]Command)

func init() {
	commands = map[string]Command{
		"ping": {
			Name:        "ping",
			Description: "Check bot latency",
			Execute: func(args []string) (string, bool) {
				start := time.Now()
				return fmt.Sprintf("%s Pong! (Latency: %v)", responsePrefix, time.Since(start)), true
			},
		},
		"echo": {
			Name:        "echo",
			Description: "Echo back your message",
			Execute: func(args []string) (string, bool) {
				if len(args) > 1 {
					return fmt.Sprintf("%s %s", responsePrefix, strings.Join(args, " ")), true
				} else {
					return "", false
				}
			},
		},
		"help": {
			Name:"Help",
			Description: "Help",
			Execute: func(args []string) (string, bool) {
				return fmt.Sprintf("%s - %s", commands[args[0]].Name, commands[args[0]].Description), true
			},
		},
	}
}

func HandleCommand(commandStr string, args []string) (string, bool) {
	commandName := strings.TrimPrefix(commandStr, prefix)
	
	command, exists := commands[commandName]
	if !exists {
		return "", false
	}
	
	response, status := command.Execute(args)
	return response, status
}
