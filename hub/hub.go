// Package hub contains the Hub class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package hub

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/svera/sackson-server/bridges"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/room"
)

// Error messages returned from hub
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
	Messages chan *interfaces.MessageFromClient

	// Registration requests
	Register chan interfaces.Client

	// Unregistration requests
	Unregister chan interfaces.Client

	// Configuration
	configuration *config.Config

	observer interfaces.Observable

	debug bool
}

var rn *rand.Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rn = rand.New(source)
}

// New returns a new Hub instance
func New(cfg *config.Config, observer interfaces.Observable, debug bool) *Hub {
	return &Hub{
		Messages:      make(chan *interfaces.MessageFromClient),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
		observer:      observer,
		debug:         debug,
	}
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	defer func() {
		for _, cl := range h.clients {
			cl.Close()
		}
	}()

	h.observer.On(room.GameStarted, func() {
		h.sendUpdatedRoomList()
	})

	for {
		select {

		case c := <-h.Register:
			h.clients = append(h.clients, c)
			h.sendUpdatedRoomList()
			if h.debug {
				log.Printf("Client added to hub, number of connected clients: %d\n", len(h.clients))
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
			h.parseMessage(m)
			break

		}
	}
}

// parseMessage distinguish the passed message between be a control message (not
// related to a particular game, but to the server) or a room one (specific to
// a particular room)
func (h *Hub) parseMessage(m *interfaces.MessageFromClient) {
	if h.isControlMessage(m) {
		h.parseControlMessage(m)
	} else {
		h.passMessageToRoom(m)
	}
}

func (h *Hub) isControlMessage(m *interfaces.MessageFromClient) bool {
	switch m.Content.Type {
	case
		interfaces.ControlMessageTypeCreateRoom,
		interfaces.ControlMessageTypeJoinRoom,
		interfaces.ControlMessageTypeTerminateRoom:
		return true
	}
	return false
}

func (h *Hub) parseControlMessage(m *interfaces.MessageFromClient) {

	switch m.Content.Type {

	case interfaces.ControlMessageTypeCreateRoom:
		var parsed interfaces.MessageCreateRoomParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if bridge, err := bridges.Create(parsed.BridgeName); err != nil {
				h.sendErrorMessage(m.Author, errors.New(InexistentBridge))
			} else {
				id := h.createRoom(bridge, m.Author)
				if response, err := h.rooms[id].AddClient(m.Author); err != nil {
					h.sendErrorMessage(m.Author, err)
				} else {
					h.broadcast(response)
				}
			}
		}

	case interfaces.ControlMessageTypeJoinRoom:
		var parsed interfaces.MessageJoinRoomParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if room, ok := h.rooms[parsed.Room]; ok {
				if response, err := room.AddClient(m.Author); err != nil {
					h.sendErrorMessage(m.Author, err)
				} else {
					h.broadcast(response)
				}
			} else {
				h.sendErrorMessage(m.Author, errors.New(InexistentRoom))
			}

		}

	case interfaces.ControlMessageTypeTerminateRoom:
		if m.Author != m.Author.Room().Owner() {
			return
		}
		h.destroyRoom(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedTerminated)
	}
}

func (h *Hub) passMessageToRoom(m *interfaces.MessageFromClient) {
	if response, err := m.Author.Room().ParseMessage(m); err != nil {
		h.sendErrorMessage(m.Author, err)
	} else {
		h.broadcast(response)
	}
}

func (h *Hub) broadcast(response map[interfaces.Client][]byte) {
	for cl, msg := range response {
		h.sendMessage(cl, msg)
	}
}

func (h *Hub) sendToAll(message []byte) {
	for _, cl := range h.clients {
		h.sendMessage(cl, message)
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
			if c.Room() != nil {
				r := c.Room()
				response := r.RemoveClient(c)
				h.broadcast(response)
				if len(r.HumanClients()) == 0 {
					h.destroyRoom(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
				}
			}
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			if h.debug {
				log.Printf("Client removed from hub, number of clients left: %d\n", len(h.clients))
			}
			break
		}
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients() int {
	return len(h.clients)
}

func (h *Hub) sendErrorMessage(author interfaces.Client, err error) {
	res := &interfaces.MessageError{
		Type:    "err",
		Content: err.Error(),
	}
	response, _ := json.Marshal(res)
	h.sendMessage(author, response)
}

func (h *Hub) createRoom(b interfaces.Bridge, owner interfaces.Client) string {
	id := h.generateID()
	h.rooms[id] = room.New(id, b, owner, h.Messages, h.Unregister, h.configuration, h.observer)

	timer := time.AfterFunc(time.Minute*h.configuration.Timeout, func() {
		if h.debug {
			log.Printf("Destroying room %s due to timeout\n", id)
		}
		h.destroyRoom(id, interfaces.ReasonRoomDestroyedTimeout)
	})
	h.rooms[id].SetTimer(timer)

	msgRoomCreated := interfaces.MessageRoomCreated{
		Type: interfaces.TypeMessageRoomCreated,
		ID:   id,
	}
	response, _ := json.Marshal(msgRoomCreated)
	h.sendMessage(owner, response)

	h.sendUpdatedRoomList()

	if h.debug {
		log.Printf("Room %s created\n", id)
	}

	return id
}

// Return a list of all rooms IDs which haven't started a game
func (h *Hub) getWaitingRoomsIds() []string {
	ids := []string{}
	for id, room := range h.rooms {
		if !room.GameStarted() {
			ids = append(ids, id)
		}
	}
	return ids
}

func (h *Hub) destroyRoom(roomID string, reasonCode string) {
	r := h.rooms[roomID]
	r.Timer().Stop()
	msg := interfaces.MessageRoomDestroyed{
		Type:   interfaces.TypeMessageRoomDestroyed,
		Reason: reasonCode,
	}
	response, _ := json.Marshal(msg)
	for _, cl := range r.Clients() {
		if cl != nil && cl.IsBot() {
			if h.debug {
				log.Printf("Bot %s destroyed", cl.Name())
			}
			cl.Close()
		} else if cl != nil {
			h.sendMessage(cl, response)
			cl.SetRoom(nil)
		}
	}

	delete(h.rooms, roomID)
	h.sendUpdatedRoomList()

	if h.debug {
		log.Printf("Room %s destroyed\n", roomID)
	}
}

func (h *Hub) sendUpdatedRoomList() {
	msgRoomList := interfaces.MessageRoomsList{
		Type:   interfaces.TypeMessageRoomsList,
		Values: h.getWaitingRoomsIds(),
	}
	response, _ := json.Marshal(msgRoomList)
	h.sendToAll(response)
}

func (h *Hub) generateID() string {
	letters := `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	var locator string
	var randomPosition int
	numberLetters := len(letters)
	for {
		for i := 0; i < 5; i++ {
			randomPosition = rn.Intn(numberLetters - 1)
			locator += string(letters[randomPosition])
		}
		if _, exists := h.rooms[locator]; !exists {
			return locator
		}
	}
}
