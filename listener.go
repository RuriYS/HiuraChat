package main

import (
	"strings"
)

var started = false
var prefix = "$"
var responsePrefix = "[BOT: Hermit]"

func listen() {
	for {
		var response Response
		err := readJSON(&response)
		if err != nil {
			logger.Error("Error reading: %v", err)
			return
		}

		if response.ConnectionId != "" {
			botId = response.ConnectionId
			if started == false {
				logger.Info("Connected as: %s (%s)", response.Name, response.ConnectionId)
				started = true
			}
			continue
		}

		if response.Message == "" || response.Sender == botId {
			continue
		}

		parts := strings.Fields(response.Message)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		if strings.HasPrefix(command, prefix) {
			if response, ok := HandleCommand(command, args); ok {
				sendMessage(response)
			}
		}

		logger.Info("%s: %s", response.SenderName, response.Message)
	}
}
