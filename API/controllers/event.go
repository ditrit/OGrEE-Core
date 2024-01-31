package controllers

import (
	"fmt"
	"net/http"
	"time"
)

func CreateEventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Simulate sending events (you can replace this with real data)
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("Event %d", i))
		time.Sleep(3 * time.Second)
		w.(http.Flusher).Flush()
	}

	// Simulate closing the connection
	fmt.Fprintf(w, "event: close")
	w.(http.Flusher).Flush()
}
