// Copyright (c) Ignite, Inc.

package main

const PORT = ":8080"

func main() {
	server := NewSSEServer()
	server.SetUp()
	go broadcast()
	server.Run()
}
