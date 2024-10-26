package main

type Message struct {
	Action string      `json:"action"`
	Data   *MessageData `json:"data,omitempty"`
}

type MessageData struct {
	Message string `json:"message"`
}

type Response struct {
	ConnectionId string `json:"connectionId,omitempty"`
	Name        string `json:"name,omitempty"`
	Message     string `json:"message,omitempty"`
	Sender      string `json:"sender,omitempty"`
	SenderName  string `json:"senderName,omitempty"`
}

type Command struct {
	Name        string
	Description string
	Execute     func(args []string) (string, bool)
}

