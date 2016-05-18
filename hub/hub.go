package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/interfaces"
)

const (
	InexistentClient  = "inexistent_client"
	OwnerNotRemovable = "owner_not_removable"
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

	// Stops hub server
	Quit chan bool

	gameBridge interfaces.Bridge
}

// New returns a new Hub instance
func New(b interfaces.Bridge) *Hub {
	return &Hub{
		Messages:   make(chan *client.Message),
		Register:   make(chan interfaces.Client),
		Unregister: make(chan interfaces.Client),
		Quit:       make(chan bool),
		clients:    []interfaces.Client{},
		gameBridge: b,
	}
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	for {
		if h.gameBridge.IsGameOver() {
			log.Println("Game ended")
			break
		}
		select {
		case c := <-h.Register:
			if err := h.addClient(c); err != nil {
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

		case <-h.Quit:
			for _, client := range h.clients {
				h.removeClient(client)
				close(client.Incoming())
			}
			return

		case m := <-h.Messages:
			h.parseMessage(m)
			break
		}
	}
}

// parseMessage distinguish the passed message between be a control message (not
// related to a particular game, but to the server) or a game one (specific to
// the game)
func (h *Hub) parseMessage(m *client.Message) {
	if h.isControlMessage(m) {
		h.parseControlMessage(m)
	} else {
		h.parseGameMessage(m)
	}
}

func (h *Hub) isControlMessage(m *client.Message) bool {
	switch m.Content.Type {
	case
		client.ControlMessageTypeAddBot,
		client.ControlMessageTypeStartGame,
		client.ControlMessageTypeKickPlayer:
		return true
	}
	return false
}

func (h *Hub) parseControlMessage(m *client.Message) {
	if !m.Author.Owner() {
		return
	}
	switch m.Content.Type {
	case client.ControlMessageTypeStartGame:
		if err := h.gameBridge.StartGame(); err == nil {
			h.broadcastUpdate()
		} else {
			h.sendErrorMessage(err, m.Author)
		}
	case client.ControlMessageTypeAddBot:
		var parsed client.AddBotMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if c, err := h.gameBridge.AddBot(parsed.BotName); err == nil {
				c.SetName(fmt.Sprintf("Player %d", h.NumberClients()+1))
				if err := h.addClient(c); err == nil {
					go c.WritePump()
					go c.ReadPump(h.Messages, h.Unregister)
				} else {
					h.sendErrorMessage(err, m.Author)
				}
			}
		}
	case client.ControlMessageTypeKickPlayer:
		var parsed client.KickPlayerMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if err := h.kickClient(parsed.PlayerNumber); err != nil {
				h.sendErrorMessage(err, m.Author)
			}
		}
	}
}

func (h *Hub) parseGameMessage(m *client.Message) {
	var err error
	var currentPlayer interfaces.Client

	if currentPlayer, err = h.currentPlayerClient(); m.Author == currentPlayer && err == nil {
		err = h.gameBridge.ParseMessage(m.Content.Type, m.Content.Params)
	}
	if err != nil {
		log.Println(err)
		h.sendErrorMessage(err, m.Author)
	} else {
		h.broadcastUpdate()
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
		msg := setOwnerMessage{
			Type: "ctl",
			Role: "mng",
		}
		response, _ := json.Marshal(msg)
		h.sendMessage(c, response)
	}
	h.sendUpdatedPlayersList()
	log.Printf("Numero de clientes: %d\n", len(h.clients))
	return nil
}

func (h *Hub) kickClient(number int) error {
	if number < 0 || number > len(h.clients) {
		return errors.New(InexistentClient)
	}
	if h.clients[number].Owner() {
		return errors.New(OwnerNotRemovable)
	}
	h.clients[number].Close()
	h.removeClient(h.clients[number])
	h.gameBridge.RemovePlayer(number)
	h.sendUpdatedPlayersList()
	return nil
}

func (h *Hub) sendUpdatedPlayersList() {
	msg := currentPlayersMessage{
		Type:   "pls",
		Values: h.clientNames(),
	}
	response, _ := json.Marshal(msg)
	for _, c := range h.clients {
		h.sendMessage(c, response)
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients() int {
	return len(h.clients)
}

func (h *Hub) sendErrorMessage(err error, author interfaces.Client) {
	res := &errorMessage{
		Type:    "err",
		Content: err.Error(),
	}
	response, _ := json.Marshal(res)
	h.sendMessage(author, response)
}
