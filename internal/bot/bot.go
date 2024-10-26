package bot

import (
	"hiurachat/internal/config"
	"hiurachat/internal/connection"
	"hiurachat/internal/handler"
	"hiurachat/internal/logger"
	"hiurachat/internal/types"
)

type Bot struct {
	conn     *connection.Client
	logger   *logger.Logger
	handler  *handler.MessageHandler
	commands map[string]types.Command
	config   *config.Config
}

func New(logger *logger.Logger, cfg *config.Config) (*Bot, error) {
	logger.Info("Initializing...")

	conn, err := connection.New(logger, cfg.WebSocket.URL)
	if err != nil {
		return nil, err
	}

	handler := handler.New(logger, cfg.Bot.Prefix, cfg.Bot.ResponsePrefix)

	bot := &Bot{
		conn:     conn,
		logger:   logger,
		handler:  handler,
		commands: make(map[string]types.Command),
		config:   cfg,
	}

	logger.Info("Loading commands")
	bot.initializeCommands()
	return bot, nil
}

func (b *Bot) Start() error {
	if err := b.conn.Connect(); err != nil {
		return err
	}

	b.logger.Info("Loading events")
	b.conn.StartHeartbeat()
	go b.handler.Listen(b.conn)
	b.conn.RequestID()

	return nil
}
