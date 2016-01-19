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
			for _, val := range h.clients {
				if val == c {
					h.removeClient(c)
					close(c.Incoming)
				}
			}
			break

		case m := <-h.Messages:
			if m.Author.Pl != h.game.CurrentPlayer() {
				break
			}
			if m.Content.Typ == "ply" {
				coords := m.Content.Det["til"]
				tl := coordsToTile(coords)
				var response []byte
				if err := h.game.PlayTile(tl); err != nil {
					res := &ErrorMessage{
						Type:    "err",
						Content: err,
					}
					response, _ = json.Marshal(res)
					h.sendMessage(m.Author, response)
				} else {
					commonMsg := CommonMessage{
						Type:  "upd",
						Board: h.boardOwnership(),
					}
					h.broadcastUpdate(commonMsg)
					h.playerUpdate(m.Author, commonMsg)
				}
			}

			break
		}
	}
}

func (h *Hub) broadcastUpdate(commonMsg CommonMessage) {
	response, _ := json.Marshal(commonMsg)
	for _, c := range h.clients {
		h.sendMessage(c, response)
	}
}

func (h *Hub) playerUpdate(c *client.Client, commonMsg CommonMessage) {
	directMsg := &DirectMessage{
		CommonMessage: commonMsg,
		Hand:          h.tilesToSlice(h.game.CurrentPlayer()),
	}
	response, _ := json.Marshal(directMsg)
	h.sendMessage(c, response)
}

func (h *Hub) tilesToSlice(pl player.Interface) []string {
	var hnd []string
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
	}
	return hnd
}

func (h *Hub) boardOwnership() map[string]string {
	cells := make(map[string]string)
	var letters = [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			cell := h.game.Board().Cell(number, letter)
			if cell.Owner().Type() == "corporation" {
				cells[strconv.Itoa(number)+letter] = cell.Owner().(*corporation.Corporation).Name()
			} else {
				cells[strconv.Itoa(number)+letter] = cell.Owner().Type()
			}
		}
	}
	fmt.Printf("%v", cells)
	return cells
}

func coordsToTile(tl string) tile.Interface {
	number, _ := strconv.Atoi(tl[:len(tl)-1])
	letter := string(tl[len(tl)-1 : len(tl)])
	return tile.New(number, letter, tile.Unincorporated{})
}

func (h *Hub) sendInitialHand() {
	for i, c := range h.clients {
		tiles := h.game.Player(i).Tiles()
		hnd := []string{}
		for _, tl := range tiles {
			hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
		}
		res := &DirectMessage{
			CommonMessage: CommonMessage{
				Type:  "ini",
				Board: map[string]string{},
			},
			Hand: h.tilesToSlice(c.Pl),
		}
		response, _ := json.Marshal(res)
		h.sendMessage(c, response)
	}
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

func (h *Hub) players() []player.Interface {
	var players []player.Interface
	for _, c := range h.clients {
		players = append(players, c.Pl)
	}
	return players
}

func (h *Hub) removeClient(c *client.Client) {
	for i := range h.clients {
		if h.clients[i] == c {
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			break
		}
	}
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
