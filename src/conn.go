package main

import (
	"code.google.com/p/go.net/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub
}

//Reads in requests from the clients
func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.h.broadcast <- message
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
func handleSocket(ws *websocket.Conn, hub *hub) {
	c := &connection{send: make(chan []byte, 256), ws: ws, h: hub}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()
	c.reader()
}
