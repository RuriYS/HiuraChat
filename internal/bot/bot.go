package bot

import (
	"hiurachat/internal/config"
	"hiurachat/internal/connection"
	"hiurachat/internal/handler"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
	"time"
)

type Bot struct {
	client   *connection.Client
	logger   *logger.Logger
	handler  *handler.MessageHandler
	commands map[string]types.Command
	config   *config.Config
	pingTime time.Time
}

func New(logger *logger.Logger, cfg *config.Config) (*Bot, error) {
	logger.Info("Initializing...")

	client, err := connection.New(logger, cfg.WebSocket.URL, nil)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		client:   client,
		logger:   logger,
		commands: make(map[string]types.Command),
		config:   cfg,
	}

	handler := handler.New(logger, cfg.Bot.Prefix, cfg.Bot.ResponsePrefix, bot)
	bot.handler = handler

	logger.Info("Loading commands")
	bot.initializeCommands()
	bot.handler.SetCommands(bot.commands)

	return bot, nil
}

func (b *Bot) GetLatency() time.Time {
	return b.pingTime
}

func (b *Bot) SetLatency(t time.Time) {
	b.pingTime = t
}

func (b *Bot) Start() error {
	if err := b.client.Connect(); err != nil {
		return err
	}

	b.logger.Info("Loading events")
	b.client.StartHeartbeat(time.Minute)
	go b.handler.Listen(b.client)
	b.client.RequestID()

	return nil
}
