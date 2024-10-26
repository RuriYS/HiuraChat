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
	conn     *connection.Client
	logger   *logger.Logger
	handler  *handler.MessageHandler
	commands map[string]types.Command
	config   *config.Config
	pingTime time.Time
}

func New(logger *logger.Logger, cfg *config.Config) (*Bot, error) {
	logger.Info("Initializing...")

	conn, err := connection.New(logger, cfg.WebSocket.URL)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		conn:     conn,
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

func (b *Bot) GetPingTime() time.Time {
	return b.pingTime
}

func (b *Bot) SetPingTime(t time.Time) {
	b.pingTime = t
}

func (b *Bot) Start() error {
	if err := b.conn.Connect(); err != nil {
		return err
	}

	b.logger.Info("Loading events")
	b.conn.StartHeartbeat(time.Minute)
	go b.handler.Listen(b.conn)
	b.conn.RequestID()

	return nil
}
