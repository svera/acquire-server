package hub

import (
	"encoding/json"
	"github.com/svera/acquire-server/bridge"
	"github.com/svera/acquire-server/client"
)

type Hub struct {
	// Registered clients
	clients []*client.Client

	// Inbound messages
	Messages chan *client.Message

	// Register requests
	Register chan *client.Client

	// Unregister requests
	Unregister chan *client.Client

	bridge *bridge.AcquireBridge
}

func New() *Hub {
	return &Hub{
		Messages:   make(chan *client.Message),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    []*client.Client{},
		bridge:     &bridge.AcquireBridge{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.clients = append(h.clients, c)
			h.bridge.AddPlayer()
			if len(h.clients) == 3 {
				h.bridge.NewGameMergeTest()
				h.broadcastUpdate()
			}
			break

		case c := <-h.Unregister:
			for _, val := range h.clients {
				if val == c {
					h.removeClient(c)
					close(c.Incoming)
				}
			}
			break

		case m := <-h.Messages:
			if m.Author != h.currentPlayerClient() {
				break
			}

			err := h.bridge.ParseMessage(m)

			if err != nil {
				res := &bridge.ErrorMessage{
					Type:    "err",
					Content: err.Error(),
				}
				response, _ := json.Marshal(res)
				h.sendMessage(m.Author, response)
			} else {
				h.broadcastUpdate()
			}
		}
	}
}

func (h *Hub) broadcastUpdate() {
	for n, c := range h.clients {
		msg := h.bridge.Status(n)
		if c == h.currentPlayerClient() {
			msg.Enabled = true
		} else {
			msg.Enabled = false
		}
		response, _ := json.Marshal(msg)
		h.sendMessage(c, response)
	}
}

func (h *Hub) currentPlayerClient() *client.Client {
	return h.clients[h.bridge.CurrentPlayerNumber()]
}

func (h *Hub) sendMessage(c *client.Client, message []byte) {
	select {
	case c.Incoming <- message:
		break

	// We can't reach the client
	default:
		close(c.Incoming)
		h.removeClient(c)
	}
}

func (h *Hub) removeClient(c *client.Client) {
	for i := range h.clients {
		if h.clients[i] == c {
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			break
		}
	}
}
