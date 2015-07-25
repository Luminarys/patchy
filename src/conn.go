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

/*func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.h.broadcast <- message
	}
	c.ws.Close()
}*/

func (c *connection) writer() {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, string(message))
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func handleSocket(ws *websocket.Conn, hub *hub) {
	c := &connection{send: make(chan []byte, 256), ws: ws, h: hub}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	c.writer()
}
