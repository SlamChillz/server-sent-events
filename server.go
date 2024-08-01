package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SSEServer struct {
	router *http.ServeMux
	server *http.Server
}

func NewSSEServer() *SSEServer {
	router := http.NewServeMux()
	return &SSEServer{
		router: router,
		server: &http.Server{
			Addr:    PORT,
			Handler: router,
		},
	}
}

func (s *SSEServer) SetUp() *SSEServer {
	s.router.Handle("/", http.FileServer(http.Dir("./static")))
	s.router.Handle("/sse/{id}", httpRequestLog(os.Stdout, setSSEHeaders(sseHandler)))
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Wait for interrupt
		<-signalChannel
		log.Println("[INFO] received interrupt signal, sse server shutting down")

		// Create timeout context to shut down the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			log.Fatalf("[FATAL] sse server shutdown failed with error: %s\n", err)
		}
	}()
	return s
}

func (s *SSEServer) Run() {
	log.Printf("[INFO] sse server is listening on port %s\n", PORT)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[FATAL] sse server exited with error: %s\n", err)
	}
}
