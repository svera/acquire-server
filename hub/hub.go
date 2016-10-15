// Package hub contains the Hub class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package hub

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	observable "github.com/GianlucaGuarini/go-observable"
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

var (
	mapLock sync.RWMutex
	rn      *rand.Rand
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

	observer *observable.Observable

	debug bool
}

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rn = rand.New(source)
}

// New returns a new Hub instance
func New(cfg *config.Config, observer *observable.Observable, debug bool) *Hub {
	h := &Hub{
		Messages:      make(chan *interfaces.MessageFromClient),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
		observer:      observer,
		debug:         debug,
	}

	h.observer.On(room.GameStarted, func() {
		message := h.createUpdatedRoomListMessage()
		for _, cl := range h.clients {
			h.sendMessage(cl, message)
		}
	})

	h.observer.On("messageCreated", func(clients []interfaces.Client, message []byte) {
		for _, cl := range clients {
			h.sendMessage(cl, message)
		}
	})

	h.observer.On("clientOut", func(r interfaces.Room) {
		if len(r.HumanClients()) == 0 {
			h.destroyRoom(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
	})

	return h
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	defer func() {
		for _, cl := range h.clients {
			cl.Close()
		}
	}()

	for {
		select {

		case c := <-h.Register:
			h.clients = append(h.clients, c)
			h.observer.Trigger("messageCreated", h.clients, h.createUpdatedRoomListMessage())
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
		h.createRoomAction(m)

	case interfaces.ControlMessageTypeJoinRoom:
		h.joinRoomAction(m)

	case interfaces.ControlMessageTypeTerminateRoom:
		h.terminateRoomAction(m)
	}
}

func (h *Hub) createRoomAction(m *interfaces.MessageFromClient) {
	var parsed interfaces.MessageCreateRoomParams
	if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if bridge, err := bridges.Create(parsed.BridgeName); err != nil {
			res := &interfaces.MessageError{
				Type:    "err",
				Content: err.Error(),
			}
			response, _ := json.Marshal(res)
			h.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
		} else {
			roomParams := map[string]interface{}{
				"playerTimeout": parsed.PlayerTimeout,
			}
			id := h.createRoom(bridge, roomParams, m.Author)
			h.rooms[id].AddHuman(m.Author)
		}
	}
}

func (h *Hub) joinRoomAction(m *interfaces.MessageFromClient) {
	var parsed interfaces.MessageJoinRoomParams
	if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if room, ok := h.rooms[parsed.Room]; ok {
			room.AddHuman(m.Author)
		} else {
			res := &interfaces.MessageError{
				Type:    "err",
				Content: InexistentRoom,
			}
			response, _ := json.Marshal(res)
			h.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
		}

	}
}

func (h *Hub) terminateRoomAction(m *interfaces.MessageFromClient) {
	if m.Author != m.Author.Room().Owner() {
		return
	}
	h.destroyRoom(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedTerminated)
}

func (h *Hub) passMessageToRoom(m *interfaces.MessageFromClient) {
	m.Author.Room().ParseMessage(m)
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
				r.RemoveClient(c)
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

func (h *Hub) createRoom(b interfaces.Bridge, roomParams map[string]interface{}, owner interfaces.Client) string {
	id := h.generateID()
	log.Printf("player timeout value: %d", roomParams["playerTimeout"].(int))
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
	h.observer.Trigger("messageCreated", []interfaces.Client{owner}, response)

	h.observer.Trigger("messageCreated", h.clients, h.createUpdatedRoomListMessage())

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

	h.expelClientsFromRoom(r, reasonCode)

	mapLock.RLock()
	log.Println("Preparado para destruir")
	delete(h.rooms, roomID)
	mapLock.RUnlock()
	h.observer.Trigger("messageCreated", h.clients, h.createUpdatedRoomListMessage())

	if h.debug {
		log.Printf("Room %s destroyed\n", roomID)
	}
}

func (h *Hub) expelClientsFromRoom(r interfaces.Room, reasonCode string) {
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
			h.observer.Trigger("messageCreated", h.clients, response)
			cl.SetRoom(nil)
			cl.StopTimer()
		}
	}
}

func (h *Hub) createUpdatedRoomListMessage() []byte {
	msgRoomList := interfaces.MessageRoomsList{
		Type:   interfaces.TypeMessageRoomsList,
		Values: h.getWaitingRoomsIds(),
	}
	response, _ := json.Marshal(msgRoomList)
	return response
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
