package hub

import (
	"encoding/json"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire-server/interfaces"
)

// Hub is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Bridge
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Hub struct {
	// Registered clients
	clients []interfaces.Client

	// Inbound messages
	Messages chan *client.Message

	// Register requests
	Register chan interfaces.Client

	// Unregister requests
	Unregister chan interfaces.Client

	gameBridge interfaces.Bridge
}

// New returns a new Hub instance
func New(b interfaces.Bridge) *Hub {
	return &Hub{
		Messages:   make(chan *client.Message),
		Register:   make(chan interfaces.Client),
		Unregister: make(chan interfaces.Client),
		clients:    []interfaces.Client{},
		gameBridge: b,
	}
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	var err error
	for {
		select {
		case c := <-h.Register:
			if err = h.addClient(c); err != nil {
				break
			}

		case c := <-h.Unregister:
			for _, val := range h.clients {
				if val == c {
					h.removeClient(c)
					close(c.Incoming())
				}
			}
			break

		case m := <-h.Messages:
			var response []byte

			if m.Content.Type == "ini" {
				if !m.Author.Owner() {
					break
				}

				if err = h.gameBridge.StartGame(); err != nil {
					break
				}
			} else if m.Content.Type == "bot" {
				if !m.Author.Owner() {
					break
				}
				if c, err := h.gameBridge.AddBot(); err == nil {
					h.addClient(c)
					go c.ReadPump(h.Messages, h.Unregister)
				} else {
					break
				}

			} else if currentPlayer, err := h.currentPlayerClient(); m.Author != currentPlayer || err != nil {
				break
			} else {
				response, err = h.gameBridge.ParseMessage(m.Content.Type, m.Content.Params)
			}

			if err != nil {
				h.sendMessage(m.Author, response)
			} else {
				h.broadcastUpdate()
			}
		}
	}
}

func (h *Hub) clientNames() []string {
	names := []string{}
	for _, c := range h.clients {
		names = append(names, c.Name())
	}
	return names
}

func (h *Hub) broadcastUpdate() {
	for n, c := range h.clients {
		response, _ := h.gameBridge.Status(n)
		h.sendMessage(c, response)
	}
}

func (h *Hub) currentPlayerClient() (interfaces.Client, error) {
	number, err := h.gameBridge.CurrentPlayerNumber()
	return h.clients[number], err
}

func (h *Hub) sendMessage(c interfaces.Client, message []byte) {
	select {
	case c.Incoming() <- message:
		break

	// We can't reach the client
	default:
		close(c.Incoming())
		h.removeClient(c)
	}
}

func (h *Hub) removeClient(c interfaces.Client) {
	for i := range h.clients {
		if h.clients[i] == c {
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			break
		}
	}
}

func (h *Hub) addClient(c interfaces.Client) error {
	if err := h.gameBridge.AddPlayer(); err != nil {
		return err
	}
	h.clients = append(h.clients, c)

	if len(h.clients) == 1 {
		c.SetOwner(true)
		msg := struct {
			Type string `json:"typ"`
			Role string `json:"rol"`
		}{
			Type: "ctl",
			Role: "mng",
		}
		response, _ := json.Marshal(msg)
		h.sendMessage(c, response)
	}
	msg := struct {
		Type   string   `json:"typ"`
		Values []string `json:"val"`
	}{
		Type:   "add",
		Values: h.clientNames(),
	}
	response, _ := json.Marshal(msg)
	for _, c := range h.clients {
		h.sendMessage(c, response)
	}
	return nil
}
