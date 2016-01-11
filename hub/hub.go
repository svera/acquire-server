package hub

import (
	"encoding/json"
	"github.com/svera/acquire"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/corporation"
	"github.com/svera/acquire/player"
	"github.com/svera/acquire/tileset"
	"strconv"
)

type Hub struct {
	// Registered clients
	clients []*client.Client

	// Inbound messages
	Broadcast chan string

	// Register requests
	Register chan *client.Client

	// Unregister requests
	Unregister chan *client.Client

	content string
}

func New() *Hub {
	return &Hub{
		Broadcast:  make(chan string),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    make([]*client.Client, 0),
		content:    "",
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.clients = append(h.clients, c)
			if len(h.clients) == 3 {
				corp1, _ := corporation.New("Corp a", 0)
				corp2, _ := corporation.New("Corp b", 0)
				corp3, _ := corporation.New("Corp c", 1)
				corp4, _ := corporation.New("Corp d", 1)
				corp5, _ := corporation.New("Corp e", 1)
				corp6, _ := corporation.New("Corp f", 2)
				corp7, _ := corporation.New("Corp g", 2)
				gm, _ := acquire.New(
					board.New(),
					h.players(),
					[7]corporation.Interface{
						corp1,
						corp2,
						corp3,
						corp4,
						corp5,
						corp6,
						corp7,
					},
					tileset.New(),
				)
				h.sendInitialHand(gm)
			}
			break

		/*case index, c := <-h.Unregister:
		_, ok := h.clients[index]
		if ok {
			delete(h.clients, index)
			close(c.Send)
		}
		break
		*/
		case m := <-h.Broadcast:
			h.content = m
			h.broadcastMessage()
			break
		}
	}
}

func (h *Hub) sendInitialHand(gm *acquire.Game) {
	for index, c := range h.clients {
		tiles := gm.Player(index).Tiles()
		coords := []string{}
		for _, tl := range tiles {
			coords = append(coords, tl.Letter()+strconv.Itoa(tl.Number()))
		}
		response, _ := json.Marshal(coords)
		select {
		case c.Send <- response:
			break

		// We can't reach the client
		default:
			close(c.Send)
			//delete(h.clients, index)
		}
	}
}

func (h *Hub) broadcastMessage() {
	for _, c := range h.clients {
		select {
		case c.Send <- []byte(h.content):
			break

		// We can't reach the client
		default:
			close(c.Send)
			//delete(h.clients, index)
		}
	}
}

func (h *Hub) players() []player.Interface {
	var players []player.Interface
	for _, c := range h.clients {
		players = append(players, c.Pl)
	}
	return players
}
