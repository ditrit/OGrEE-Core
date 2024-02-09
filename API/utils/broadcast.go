package utils

import "context"

// BroadcastServer attaches to a channel and notify all listeners
// every time new data arrives from the channel
// Used for the SSE stream

type BroadcastServer interface {
	Subscribe() <-chan string
	CancelSubscription(<-chan string)
}
type broadcastServer struct {
	source         <-chan string
	listeners      []chan string
	addListener    chan chan string
	removeListener chan (<-chan string)
}

func (s *broadcastServer) Subscribe() <-chan string {
	newListener := make(chan string)
	s.addListener <- newListener
	return newListener
}

func (s *broadcastServer) CancelSubscription(channel <-chan string) {
	s.removeListener <- channel
}

func NewBroadcastServer(ctx context.Context, source <-chan string) BroadcastServer {
	service := &broadcastServer{
		source:         source,
		listeners:      make([]chan string, 0),
		addListener:    make(chan chan string),
		removeListener: make(chan (<-chan string)),
	}
	go service.serve(ctx)
	return service
}

func (s *broadcastServer) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, ch := range s.listeners {
				if ch == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = s.listeners[:len(s.listeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.source:
			if !ok {
				return
			}
			for _, listener := range s.listeners {
				if listener != nil {
					select {
					case listener <- val:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}
