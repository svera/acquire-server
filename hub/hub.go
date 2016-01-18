package hub

import (
	"encoding/json"
	"fmt"
	"github.com/svera/acquire"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/corporation"
	"github.com/svera/acquire/player"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
	"strconv"
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

	game *acquire.Game
}

func New() *Hub {
	return &Hub{
		Messages:   make(chan *client.Message),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    []*client.Client{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.clients = append(h.clients, c)
			if len(h.clients) == 3 {
				h.newGame()
				h.sendInitialHand()
			}
			break

		case c := <-h.Unregister:
			for i, val := range h.clients {
				if val == c {
					h.removeClient(i)
					close(c.Send)
				}
			}
			break

		case m := <-h.Messages:
			if m.Author.Pl != h.game.CurrentPlayer() {
				fmt.Println("Player not in turn")
			} else {
				fmt.Println("Player in turn")
				if m.Content.Typ == "ply" {
					coords := m.Content.Det["til"]
					tl := coordsToTile(coords)
					if err := h.game.PlayTile(tl); err != nil {

					} else {
						res := &Message{
							Result: "ok",
							Type:   "upd",
							Board: map[string]string{
								coords: tl.Owner().Type(),
							},
							Hand: h.tilesToSlice(h.game.CurrentPlayer()),
						}
						for i, c := range h.clients {
							response, _ := json.Marshal(res)
							select {
							case c.Send <- response:
								break

							// We can't reach the client
							default:
								close(c.Send)
								h.removeClient(i)
							}
						}
					}

					//fmt.Println(h.game.StatusName())
				}
			}

			//fmt.Println(m)
			break
		}
	}
}

func (h *Hub) tilesToSlice(pl player.Interface) []string {
	var hnd []string
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
	}
	return hnd
}

func coordsToTile(tl string) tile.Interface {
	number, _ := strconv.Atoi(string(tl[0]))
	letter := string(tl[1:len(tl)])
	return tile.New(number, letter, tile.Unincorporated{})
}

func (h *Hub) sendInitialHand() {
	for i, c := range h.clients {
		tiles := h.game.Player(i).Tiles()
		hnd := []string{}
		for _, tl := range tiles {
			hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
		}
		res := &Message{
			Result: "ok",
			Type:   "ini",
			Board:  map[string]string{},
			Hand:   h.tilesToSlice(c.Pl),
		}
		response, _ := json.Marshal(res)
		select {
		case c.Send <- response:
			break

		// We can't reach the client
		default:
			close(c.Send)
			h.removeClient(i)
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

func (h *Hub) removeClient(i int) {
	h.clients = append(h.clients[:i], h.clients[i+1:]...)
}

func (h *Hub) newGame() {
	corp1, _ := corporation.New("Corp a", 0)
	corp2, _ := corporation.New("Corp b", 0)
	corp3, _ := corporation.New("Corp c", 1)
	corp4, _ := corporation.New("Corp d", 1)
	corp5, _ := corporation.New("Corp e", 1)
	corp6, _ := corporation.New("Corp f", 2)
	corp7, _ := corporation.New("Corp g", 2)
	h.game, _ = acquire.New(
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
}
