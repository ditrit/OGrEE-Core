package controllers

import (
	"context"
	"fmt"
	"net/http"
	u "p3/utils"
)

var eventNotifier chan string
var broadcaster u.BroadcastServer

func init() {
	// Create channel to send events to broadcaster
	ctx, _ := context.WithCancel(context.Background())
	eventNotifier = make(chan string)
	broadcaster = u.NewBroadcastServer(ctx, eventNotifier)
}

// swagger:operation GET /api/events Events CreateEventStream
// Get real-time notifications (SSE stream)
// Opens a SSE stream with the caller where the API will send a new event (message in JSON format)
// every time a modify or delete of any object succeeds. Also applies to create layer.
// ---
// security:
// - bearer: []
// produces:
// - text/event-stream
//
// responses:
//		'200':
//			description: 'Successfully established stream, keep it open.'

func CreateEventStream(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateEventStream ")
	fmt.Println("******************************************************")
	// Configure SSE stream
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Subscribe to broadcaster to receive events from entity controller
	listener := broadcaster.Subscribe()
	for str := range listener {
		// New event receive, send it
		fmt.Fprintf(w, "data: %v\n", str)
		w.(http.Flusher).Flush()
	}

	// Close the connection if not listening anymore
	fmt.Fprintf(w, "event: close")
	w.(http.Flusher).Flush()
}
