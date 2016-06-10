// Package hub contains the Hub class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package hub

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/svera/tbg-server/bridges"
	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/config"
	"github.com/svera/tbg-server/interfaces"
	"github.com/svera/tbg-server/room"
)

const (
	InexistentClient  = "inexistent_client"
	OwnerNotRemovable = "owner_not_removable"
	Forbidden         = "forbidden"
	InexistentRoom    = "inexistent_room"
	InexistentBridge  = "inexistent_bridge"
)

// Hub is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Bridge
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Hub struct {
	// Registered clients
	clients []interfaces.Client

	rooms map[string]interfaces.Room

	// Inbound messages
	Messages chan *interfaces.ClientMessage

	// Registration requests
	Register chan interfaces.Client

	// Unregistration requests
	Unregister chan interfaces.Client

	// Configuration
	configuration *config.Config
}

// New returns a new Hub instance
func New(cfg *config.Config) *Hub {
	return &Hub{
		Messages:      make(chan *interfaces.ClientMessage),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
	}
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {

	for {
		select {

		case c := <-h.Register:
			h.clients = append(h.clients, c)

		case c := <-h.Unregister:
			for _, val := range h.clients {
				if val == c {
					h.removeClient(c)
					close(c.Incoming())
				}
			}
			break

		case m := <-h.Messages:
			h.parseMessage(m)
			break

		}
	}
}

// parseMessage distinguish the passed message between be a control message (not
// related to a particular game, but to the server) or a game one (specific to
// the game)
func (h *Hub) parseMessage(m *interfaces.ClientMessage) {
	if h.isControlMessage(m) {
		h.parseControlMessage(m)
	} else {
		if room, ok := h.rooms[m.Content.Room]; ok {
			if err := room.ParseMessage(m); err != nil {
				h.sendErrorMessage(err, m.Author)
			} else {
				h.broadcastUpdate()
			}
		} else {
			h.sendErrorMessage(errors.New(InexistentRoom), m.Author)
		}
	}
}

func (h *Hub) isControlMessage(m *interfaces.ClientMessage) bool {
	switch m.Content.Type {
	case
		client.ControlMessageTypeCreateRoom,
		client.ControlMessageTypeTerminateRoom:
		return true
	}
	return false
}

func (h *Hub) parseControlMessage(m *interfaces.ClientMessage) {

	switch m.Content.Type {

	case client.ControlMessageTypeCreateRoom:
		var parsed client.CreateRoomMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if bridge, err := bridges.Create(parsed.BridgeName); err != nil {
				log.Println(parsed.BridgeName)
				h.sendErrorMessage(errors.New(InexistentBridge), m.Author)
			} else {
				id := h.createRoom(bridge, h.configuration)
				msg := newRoomCreatedMessage{
					Type: "new",
					ID:   id,
				}
				response, _ := json.Marshal(msg)
				h.sendMessage(m.Author, response)
				h.sendUpdatedPlayersList()
			}
		}

		/*
			case client.ControlMessageTypeKickPlayer:
				var parsed client.KickPlayerMessageParams
				if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
					if err := h.kickClient(parsed.PlayerNumber); err != nil {
						h.sendErrorMessage(err, m.Author)
					}
				}

			case client.ControlMessageTypePlayerQuits:
				if err := h.quitClient(m.Author); err != nil {
					h.sendErrorMessage(err, m.Author)
				}
			case client.ControlMessageTypeTerminateGame:
				if err := h.terminateGame(m.Author); err != nil {
					h.sendErrorMessage(err, m.Author)
				}
		*/
	}
}

func (h *Hub) clientNames() []string {
	names := []string{}
	for _, c := range h.clients {
		if c != nil {
			names = append(names, c.Name())
		}
	}
	return names
}

func (h *Hub) broadcastUpdate() {
	for n, c := range h.clients {
		if c != nil {
			if c.IsBot() && c.Room().IsGameOver() {
				continue
			}
			response, _ := c.Room().Status(n)
			h.sendMessage(c, response)
		}
	}
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

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (h *Hub) removeClient(c interfaces.Client) {
	for i := range h.clients {
		if h.clients[i] == c {
			if c.Room().GameStarted() {
				h.clients[i] = nil
				//h.gameBridge.DeactivatePlayer(i)
				h.broadcastUpdate()
			} else {
				h.clients = append(h.clients[:i], h.clients[i+1:]...)
				//h.gameBridge.RemovePlayer(i)
				h.sendUpdatedPlayersList()
			}
			log.Printf("Cliente eliminado, Numero de clientes: %d\n", len(h.clients))
			return
		}
	}
}

func (h *Hub) sendUpdatedPlayersList() {
	msg := currentPlayersMessage{
		Type:   "pls",
		Values: h.clientNames(),
	}
	response, _ := json.Marshal(msg)
	for _, c := range h.clients {
		if c != nil {
			log.Printf("sent message %s", response)
			h.sendMessage(c, response)
		}
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

func (h *Hub) createRoom(b interfaces.Bridge, cfg *config.Config) string {
	id := generateID()
	h.rooms[id] = room.New(b, cfg)
	return id
}

// TODO Implement proper random string generator
func generateID() string {
	return "a"
}
