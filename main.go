// Copyright (c) Ignite, Inc.

package main

const PORT = ":8080"

func main() {
	// Create server
	server := NewSSEServer()
	// Configure and start server
	server.SetUp().Run()
}
