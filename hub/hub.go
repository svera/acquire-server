package hub

import (
	"github.com/svera/acquire-server/client"
)

type Hub struct {
	// Registered clients
	clients map[*client.Client]bool

	// Inbound messages
	Broadcast chan string

	// Register requests
	Register chan *client.Client

	// Unregister requests
	Unregister chan *client.Client

	content string
}

func New() Hub {
	return Hub{
		Broadcast:  make(chan string),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    make(map[*client.Client]bool),
		content:    "",
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.clients[c] = true
			c.Send <- []byte(h.content)
			break

		case c := <-h.Unregister:
			_, ok := h.clients[c]
			if ok {
				delete(h.clients, c)
				close(c.Send)
			}
			break

		case m := <-h.Broadcast:
			h.content = m
			h.broadcastMessage()
			break
		}
	}
}

func (h *Hub) broadcastMessage() {
	for c := range h.clients {
		select {
		case c.Send <- []byte(h.content):
			break

		// We can't reach the client
		default:
			close(c.Send)
			delete(h.clients, c)
		}
	}
}
