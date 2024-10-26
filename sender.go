package main

func sendMessage(message string) error {
	msg := Message{
		Action: "sendMessage",
		Data: &MessageData{
			Message: message,
		},
	}
	return writeJSON(msg)
}
