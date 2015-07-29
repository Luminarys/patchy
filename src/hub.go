package main

import "encoding/json"

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func newHub() *hub {
	return &hub{
		broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			go func() {
				msg := map[string]string{"cmd": "ljoin"}
				jsonMsg, _ := json.Marshal(msg)
				h.broadcast <- []byte(jsonMsg)
			}()
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				go func() {
					msg := map[string]string{"cmd": "lleave"}
					jsonMsg, _ := json.Marshal(msg)
					h.broadcast <- []byte(jsonMsg)
				}()
			}
		//Global broadcasts
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}
