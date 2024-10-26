package main

import (
	"log"

	"hiurachat/internal/bot"
	"hiurachat/internal/logger"
)

func main() {
	l := logger.NewLogger()
	if l == nil {
		log.Fatal("Failed to initialize logger")
	}
	defer l.Close()

	l.SetLogLevel(logger.DEBUG)

	bot, err := bot.New(l)
	if err != nil {
		l.Error("Failed to initialize bot: %v", err)
		return
	}

	if err := bot.Start(); err != nil {
		l.Error("Bot failed to start: %v", err)
		return
	}

	select {}
}
