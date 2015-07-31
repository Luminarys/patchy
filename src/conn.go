package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub

	// The channel for interfacing with the song/mpd handler
	utaChan chan string

	// The channel for passing requests
	requests chan *request
}

//Reads in requests from the clients and sends them to the song handler
func (c *connection) reader() {
	var msg string
	for {
		if err := websocket.Message.Receive(c.ws, &msg); err != nil {
			break
		}
		var req map[string]string
		if err := json.Unmarshal([]byte(msg), &req); err != nil {
			fmt.Println("Error, couldn't unmarshal client request")
		} else {
			fmt.Println("Received: ", req)
			if req["cmd"] == "req" {
				c.requests <- &request{Title: req["Title"], Album: req["Album"], Artist: req["Artist"]}
			}
		}
	}
	c.ws.Close()
}

//Sends broadcasts to clients
func (c *connection) writer() {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, string(message))
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

//Socket handler -- Creates a new connection for each client
func handleSocket(ws *websocket.Conn, hub *hub, utaChan chan string, requests chan *request) {
	c := &connection{send: make(chan []byte, 256), ws: ws, h: hub, utaChan: utaChan, requests: requests}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()
	c.reader()
}
