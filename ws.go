package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

var conn *websocket.Conn
var botId string
var connMutex sync.Mutex

func connect() error {
	url := "wss://gddr51pi43.execute-api.ap-southeast-1.amazonaws.com/dev"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	conn = c
	return nil
}

func requestId() error {
	msg := Message{
		Action: "getId",
	}
	return writeJSON(msg)
}

func writeJSON(v interface{}) error {
	connMutex.Lock()
	defer connMutex.Unlock()
	return conn.WriteJSON(v)
}

func readJSON(v interface{}) error {
	connMutex.Lock()
	defer connMutex.Unlock()
	return conn.ReadJSON(v)
}