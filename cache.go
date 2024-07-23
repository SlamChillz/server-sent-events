package main

import "time"

var clients = make(map[chan string]<-chan struct{})

// broadcast sends message to all connected client
func broadcast() {
	for {
		clientCount := len(clients)
		if clientCount == 0 {
			continue
		}
		eventData := `{"count": "one"}`
		for client := range clients {
			client <- eventData
			time.Sleep(100 * time.Millisecond)
		}
	}
}
