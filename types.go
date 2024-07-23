package main

import (
	"net/http"
)

// logResponseWriter holds information about the stream for logging
type logResponseWriter struct {
	writer http.ResponseWriter
	status int
	length int
}

// Header implements the http.ResponseWriter interface Header method
func (w *logResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// WriteHeader implements the http.ResponseWriter interface WriteHeader method
func (w *logResponseWriter) WriteHeader(s int) {
	w.status = s
	w.writer.WriteHeader(s)
}

func (w *logResponseWriter) Flush() {
	w.writer.(http.Flusher).Flush()
}

// Write implements the http.ResponseWriter interface Write method
func (w *logResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.length = len(b)
	n, err := w.writer.Write(b)
	w.Flush()
	return n, err
}
