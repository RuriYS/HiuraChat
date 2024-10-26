package main

import (
	"log"

	"hiurachat/internal/bot"
	"hiurachat/internal/config"
	"hiurachat/internal/logger"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	l := logger.NewLogger()
	if l == nil {
		log.Fatal("Failed to initialize logger")
	}
	defer l.Close()

	switch cfg.Logger.Level {
	case "debug":
		l.SetLogLevel(logger.DEBUG)
	case "info":
		l.SetLogLevel(logger.INFO)
	case "warn":
		l.SetLogLevel(logger.WARN)
	case "error":
		l.SetLogLevel(logger.ERROR)
	}
	l.SetUseColors(cfg.Logger.UseColors)

	bot, err := bot.New(l, cfg)
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
