package controllers

import (
	"fmt"
	"net/http"
	"time"
)

var eventNotifier chan string

func init() {
	eventNotifier = make(chan string)
}

func CreateEventStream(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateEventStream ")
	fmt.Println("******************************************************")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Simulate sending events (you can replace this with real data)
	for i := 0; i < 1; i++ {
		fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("Event %d", i))
		time.Sleep(3 * time.Second)
		w.(http.Flusher).Flush()
	}

	// done := make(chan bool)

	// go func() {
	// 	if msg, ok := <-output; ok {
	// 		c.SSEvent("msg", string(msg))
	// 		return true
	// 	}
	// 	return false
	// 	done <- true
	// }()
	for str := range eventNotifier {
		println("GOT IT")
		fmt.Println(str)
		fmt.Fprintf(w, "data: %v\n", str)
		w.(http.Flusher).Flush()
	}

	// <-done

	println("##END")
	// Simulate closing the connection
	fmt.Fprintf(w, "event: close")
	w.(http.Flusher).Flush()
}
