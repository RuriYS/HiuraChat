package bot

import (
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
}

func New(logger *logger.Logger) (*Bot, error) {
    conn, err := connection.New(logger)
    if err != nil {
        return nil, err
    }

    handler := handler.New(logger)
    
    bot := &Bot{
        conn:     conn,
        logger:   logger,
        handler:  handler,
        commands: make(map[string]types.Command),
    }
    
    bot.initializeCommands()
    return bot, nil
}

func (b *Bot) Start() error {
    if err := b.conn.Connect(); err != nil {
        return err
    }

    b.conn.StartHeartbeat()
    go b.handler.Listen(b.conn)
    
    return nil
}