package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/svera/acquire"
	"github.com/svera/acquire-server/client"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/corporation"
	"github.com/svera/acquire/fsm"
	"github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
	"strconv"
	"strings"
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
				h.newGameTiedMergeTest()
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
			var err error
			var ok bool

			if m.Author.Pl != h.game.CurrentPlayer() {
				break
			}

			// TODO: look for a type switch for this case
			switch m.Content.Type {
			case "ply":
				if params, ok := m.Content.Params.(client.PlayTileMessageParams); ok {
					err = h.playTile(params, m.Author)
				}
			case "ncp":
				if params, ok := m.Content.Params.(client.NewCorpMessageParams); ok {
					err = h.foundCorporation(params, m.Author)
				}
			case "buy":
				if params, ok := m.Content.Params.(client.BuyMessageParams); ok {
					err = h.buyStock(params, m.Author)
				}
			case "sel":
				if params, ok := m.Content.Params.(client.SellTradeMessageParams); ok {
					err = h.sellTrade(params, m.Author)
				}
			case "unt":
				if params, ok := m.Content.Params.(client.UntieMergeMessageParams); ok {
					err = h.untieMerge(params, m.Author)
				}
			default:
				err = errors.New("Message parsing error")
			}

			if !ok {
				fmt.Println("%v", m.Content.Params)
				err = errors.New("Message parsing error")
			}

			if err != nil {
				res := &ErrorMessage{
					Type:    "err",
					Content: err.Error(),
				}
				response, _ := json.Marshal(res)
				h.sendMessage(m.Author, response)
			}
			break
		}
	}
}

func (h *Hub) playTile(params client.PlayTileMessageParams, c *client.Client) error {
	var err error

	if tl, err := coordsToTile(params.Tile); err == nil {
		if err := h.game.PlayTile(tl); err == nil {
			h.broadcastUpdate()
			h.playerUpdate(c)
			return nil
		}
	}
	return err
}

func (h *Hub) foundCorporation(params client.NewCorpMessageParams, c *client.Client) error {
	var err error

	if corp, err := h.findCorpByName(params.Corporation); err == nil {
		if err := h.game.FoundCorporation(corp); err == nil {
			h.broadcastUpdate()
			h.playerUpdate(c)
			return nil
		}
	}
	return err
}

func (h *Hub) buyStock(params client.BuyMessageParams, c *client.Client) error {
	var err error
	buy := map[interfaces.Corporation]int{}

	for corpName, amount := range params.Corporations {
		if corp, err := h.findCorpByName(corpName); err == nil {
			buy[corp] = amount
		} else {
			return err
		}
	}

	if err := h.game.BuyStock(buy); err == nil {
		h.broadcastUpdate()
		h.playerUpdate(h.getCurrentPlayerClient())
		return nil
	}
	return err
}

func (h *Hub) sellTrade(params client.SellTradeMessageParams, c *client.Client) error {
	var err error
	sell := map[interfaces.Corporation]int{}
	trade := map[interfaces.Corporation]int{}

	for corpName, operation := range params.Corporations {
		if corp, err := h.findCorpByName(corpName); err == nil {
			sell[corp] = operation.Sell
			trade[corp] = operation.Trade
		} else {
			return err
		}
	}

	if err := h.game.SellTrade(sell, trade); err == nil {
		h.broadcastUpdate()
		h.playerUpdate(h.getCurrentPlayerClient())
		return nil
	}
	return err
}

func (h *Hub) untieMerge(params client.UntieMergeMessageParams, c *client.Client) error {
	var err error

	if corp, err := h.findCorpByName(params.Corporation); err == nil {
		if err := h.game.UntieMerge(corp); err == nil {
			h.broadcastUpdate()
			h.playerUpdate(c)
			return nil
		}
	}
	return err
}

func (h *Hub) broadcastUpdate() {
	commonMsg := CommonMessage{
		Type:  "upd",
		Board: h.boardOwnership(),
	}
	for _, c := range h.clients {
		if c.Pl == h.game.CurrentPlayer() {
			commonMsg.Enabled = true
		} else {
			commonMsg.Enabled = false
		}
		response, _ := json.Marshal(commonMsg)
		h.sendMessage(c, response)
	}
}

func (h *Hub) getCurrentPlayerClient() *client.Client {
	cl := &client.Client{}
	for _, cl = range h.clients {
		if cl.Pl == h.game.CurrentPlayer() {
			break
		}
	}
	return cl
}

func (h *Hub) playerUpdate(c *client.Client) {
	directMsg := &DirectMessage{
		Type:          "dir",
		Hand:          h.tilesToSlice(c.Pl),
		State:         h.game.GameStateName(),
		InactiveCorps: corpNames(h.game.InactiveCorporations()),
		ActiveCorps:   corpNames(h.game.ActiveCorporations()),
		TiedCorps:     corpNames(h.game.TiedCorps()),
		Shares:        h.mapShares(c.Pl),
	}
	response, _ := json.Marshal(directMsg)
	h.sendMessage(c, response)
}

func (h *Hub) tilesToSlice(pl interfaces.Player) []string {
	var hnd []string
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
	}
	return hnd
}

func corpNames(corps []interfaces.Corporation) []string {
	names := []string{}
	for _, corp := range corps {
		names = append(names, corp.Name())
	}
	return names
}

func (h *Hub) mapShares(pl interfaces.Player) map[string]int {
	corps := map[string]int{}
	for _, c := range h.game.DefunctCorporations() {
		if amount := pl.Shares(c); amount > 0 {
			corps[strings.ToLower(c.Name())] = amount
		}
	}
	return corps
}

func (h *Hub) findCorpByName(name string) (interfaces.Corporation, error) {
	for _, corp := range h.game.Corporations() {
		if strings.ToLower(corp.Name()) == name {
			return corp, nil
		}
	}
	return &corporation.Corporation{}, errors.New("corporation not found")
}

func (h *Hub) boardOwnership() map[string]string {
	cells := make(map[string]string)
	var letters = [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			cell := h.game.Board().Cell(number, letter)
			if cell.Type() == "corporation" {
				cells[strconv.Itoa(number)+letter] = strings.ToLower(cell.(*corporation.Corporation).Name())
			} else {
				cells[strconv.Itoa(number)+letter] = cell.Type()
			}
		}
	}

	return cells
}

func coordsToTile(tl string) (interfaces.Tile, error) {
	if len(tl) < 2 {
		return &tile.Tile{}, errors.New("Not a valid tile")
	}
	number, _ := strconv.Atoi(tl[:len(tl)-1])
	letter := string(tl[len(tl)-1 : len(tl)])
	return tile.New(number, letter), nil
}

func (h *Hub) sendInitialHand() {
	for i, c := range h.clients {
		tiles := h.game.Player(i).Tiles()
		hnd := []string{}
		for _, tl := range tiles {
			hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
		}
		h.playerUpdate(c)
	}
	h.broadcastUpdate()
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

func (h *Hub) players() []interfaces.Player {
	var players []interfaces.Player
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
	h.game, _ = acquire.New(
		board.New(),
		h.players(),
		createCorporations(),
		tileset.New(),
		&fsm.PlayTile{},
	)
}

func createCorporations() [7]interfaces.Corporation {
	var corps [7]interfaces.Corporation
	corpsParams := [7]map[string]int{
		map[string]int{"Sackson": 0},
		map[string]int{"Zeta": 0},
		map[string]int{"Hydra": 1},
		map[string]int{"Fusion": 1},
		map[string]int{"America": 1},
		map[string]int{"Phoenix": 2},
		map[string]int{"Quantum": 2},
	}

	for i, corpData := range corpsParams {
		for corpName, corpClass := range corpData {
			if corp, err := corporation.New(corpName, corpClass); err == nil {
				corps[i] = corp
			} else {
				panic(err)
			}
		}
	}
	return corps
}

func (h *Hub) newGameMergeTest() {
	bd := board.New()
	ts := tileset.NewStub()
	corps := createCorporations()
	tiles := []interfaces.Tile{
		tile.New(5, "E"),
		tile.New(6, "E"),
	}
	tiles2 := []interfaces.Tile{
		tile.New(8, "E"),
		tile.New(9, "E"),
		tile.New(10, "E"),
	}

	ts.DiscardTile(tiles[0])
	ts.DiscardTile(tiles[1])
	ts.DiscardTile(tiles2[0])
	ts.DiscardTile(tiles2[1])
	ts.DiscardTile(tiles2[2])
	bd.SetOwner(corps[0], tiles)
	bd.SetOwner(corps[1], tiles2)
	corps[0].Grow(2)
	corps[1].Grow(3)

	h.game, _ = acquire.New(
		bd,
		h.players(),
		corps,
		tileset.New(),
		&fsm.PlayTile{},
	)

	h.players()[0].DiscardTile(h.players()[0].Tiles()[0])
	h.players()[0].PickTile(tile.New(7, "E"))
	h.players()[0].AddShares(corps[0], 5)
	h.players()[1].AddShares(corps[0], 5)
}

func (h *Hub) newGameTiedMergeTest() {
	bd := board.New()
	ts := tileset.NewStub()
	corps := createCorporations()
	tiles := []interfaces.Tile{
		tile.New(4, "E"),
		tile.New(5, "E"),
		tile.New(6, "E"),
	}
	tiles2 := []interfaces.Tile{
		tile.New(8, "E"),
		tile.New(9, "E"),
		tile.New(10, "E"),
	}

	ts.DiscardTile(tiles[0])
	ts.DiscardTile(tiles[1])
	ts.DiscardTile(tiles[2])
	ts.DiscardTile(tiles2[0])
	ts.DiscardTile(tiles2[1])
	ts.DiscardTile(tiles2[2])
	bd.SetOwner(corps[0], tiles)
	bd.SetOwner(corps[1], tiles2)
	corps[0].Grow(3)
	corps[1].Grow(3)

	h.game, _ = acquire.New(
		bd,
		h.players(),
		corps,
		tileset.New(),
		&fsm.PlayTile{},
	)

	h.players()[0].DiscardTile(h.players()[0].Tiles()[0])
	h.players()[0].PickTile(tile.New(7, "E"))
	h.players()[0].AddShares(corps[0], 5)
	h.players()[1].AddShares(corps[0], 5)
	h.players()[0].AddShares(corps[1], 3)
	h.players()[1].AddShares(corps[1], 3)
}
