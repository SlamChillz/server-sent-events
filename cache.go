package main

var clients = make(map[string]chan<- string)

// broadcast sends message to all connected client
//func broadcast() {
//	for {
//		clientCount := len(clients)
//		if clientCount == 0 {
//			continue
//		}
//		eventData := `{"count": "one"}`
//		for client := range clients {
//			client <- eventData
//			time.Sleep(100 * time.Millisecond)
//		}
//	}
//}
