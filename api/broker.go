package api

import (
	"fmt"
	"net/http"
)

type Broker struct {
	// channels for con/dcon
	Connecting    chan chan []byte
	Disconnecting chan chan []byte

	// channel for broadcast
	Notifier chan []byte

	// map of clients
	clients map[chan []byte]bool
}

func NewBroker() *Broker {
	return &Broker{
		Connecting:    make(chan chan []byte),
		Disconnecting: make(chan chan []byte),
		Notifier:      make(chan []byte),
		clients:       make(map[chan []byte]bool),
	}
}

func (b *Broker) Run() {
	for {
		select {
		case s := <-b.Connecting:
			b.clients[s] = true
		case s := <-b.Disconnecting:
			delete(b.clients, s)
		case msg := <-b.Notifier:
			for s := range b.clients {
				s <- msg
			}
		}
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// type assertion for http.Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// connect / disconnect
	clientChan := make(chan []byte)
	b.Connecting <- clientChan

	defer func() {
		b.Disconnecting <- clientChan
	}()

	// loop
	for {
		select {
		// if done: get out
		case <-r.Context().Done():
			return
		// if message sent from client channel: flush to client
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
