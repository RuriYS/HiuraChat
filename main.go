package main

import (
	"log"
)

var logger *Logger

func main() {
	// Initialize logger
	logger = getLogger()
	if logger == nil {
		log.Fatal("Failed to initialize logger")
	}
	defer logger.Close()

	logger.SetLogLevel(DEBUG)

	logger.Info("Starting...")

	err := connect()
	if err != nil {
		logger.Error("Failed to connect: %v", err)
		return
	}
	defer conn.Close()

	err = requestId()
	if err != nil {
		logger.Error("Failed to request ID: %v", err)
		return
	}

	startHeartbeat()
	go listen()
	
	select{}
}