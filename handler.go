package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const httpRequestLogFormat = "%v %s %s \"%s %s %s\" %d %d \"%s\" %v\n"

// httpRequestLog a middleware that logs to a given Writer
func httpRequestLog(out io.Writer, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lrw logResponseWriter
		lrw.writer = w
		defer func(start time.Time) {
			status := lrw.status
			length := lrw.length
			end := time.Now()
			duration := end.Sub(start)
			_, err := fmt.Fprintf(out, httpRequestLogFormat,
				end.Format(time.RFC3339),
				r.Host, r.RemoteAddr, r.Method, r.URL.Path, r.Proto,
				status, length, r.UserAgent(), duration)
			if err != nil {
				log.Println("[httpRequestLog]", err)
			}
		}(time.Now())
		h(&lrw, r)
	}
}

// setSSEHeaders adds the server sent event headers to every valid request
func setSSEHeaders(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		h(w, r)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	done := make(chan bool)
	closeStreamConsumer := make(chan bool)
	streamCreated := make(chan bool)
	// Listen for a close event from client
	//id := r.URL.Query().Get("id")
	id := "mendy"
	reqChan := r.Context().Done()
	eventChan := make(chan string)
	clients[id] = eventChan

	defer func() {
		delete(clients, id)
		close(eventChan)
	}()

	go publish(id, streamCreated)

	go func() {
		for {
			select {
			case <-reqChan:
				log.Printf("[INFO] %v client closed connection", r.RemoteAddr)
				done <- true
				return
			default:
				time.Sleep(600 * time.Millisecond)
				data, more := <-eventChan
				if !more {
					continue
				}
				//fmt.Println(len(data))
				_, err := fmt.Fprintf(w, "data: %v\n\n", fmt.Sprintf("[%v]: %v", time.Now().Format(time.RFC3339Nano), data))
				if err != nil {
					log.Printf("[ERROR] error sending event to client, %v: %v", r.RemoteAddr, err)
					return
				}
				log.Printf("[INFO] %v client got event: %v", r.RemoteAddr, data)
			}
		}
	}()
	<-streamCreated
	defer close(streamCreated)
	streamConsumer := NewStreamConsumer(id, done, closeStreamConsumer)
	streamConsumerChanClose := streamConsumer.Consumer.NotifyClose()
	defer func() {
		event := <-streamConsumerChanClose
		log.Printf("[INFO] Consumer: %s closed on the stream: %s, reason: %s \n", event.Name, event.StreamName, event.Reason)
	}()
	<-closeStreamConsumer
	defer close(closeStreamConsumer)
}
