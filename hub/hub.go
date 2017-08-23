package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

var (
	mutex sync.RWMutex
	rn    *rand.Rand
	wg    sync.WaitGroup
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
	Messages chan *interfaces.IncomingMessage

	// Registration requests
	Register chan interfaces.Client

	// Unregistration requests
	Unregister chan interfaces.Client

	// Configuration
	configuration *config.Config

	observer interfaces.Observer
}

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rn = rand.New(source)
}

// New returns a new Hub instance
func New(cfg *config.Config, obs interfaces.Observer) *Hub {
	h := &Hub{
		Messages:      make(chan *interfaces.IncomingMessage),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
		observer:      obs,
	}

	h.registerEvents()

	return h
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	for {
		select {

		case cl := <-h.Register:
			mutex.Lock()
			h.clients = append(h.clients, cl)
			mutex.Unlock()
			cl.SetName(fmt.Sprintf("Player %d", h.NumberClients()))
			h.observer.Trigger("messageCreated", []interfaces.Client{cl}, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
			if h.configuration.Debug {
				log.Printf("Client added to hub, number of connected clients: %d\n", len(h.clients))
			}

		case cl := <-h.Unregister:
			for _, val := range h.clients {
				if val == cl {
					wg.Wait()
					h.removeClient(cl)
					break
				}
			}

		case m := <-h.Messages:
			h.parseMessage(m)

		}
	}
}

// parseMessage distinguish the passed message between be a control message (not
// related to a particular game, but to the server) or a room one (specific to
// a particular room)
func (h *Hub) parseMessage(m *interfaces.IncomingMessage) {
	if h.isControlMessage(m) {
		h.parseControlMessage(m)
	} else {
		h.passMessageToRoom(m)
	}
}

func (h *Hub) isControlMessage(m *interfaces.IncomingMessage) bool {
	switch m.Content.Type {
	case
		interfaces.ControlMessageTypeCreateRoom,
		interfaces.ControlMessageTypeJoinRoom,
		interfaces.ControlMessageTypeTerminateRoom:
		return true
	}
	return false
}

func (h *Hub) parseControlMessage(m *interfaces.IncomingMessage) {
	var err error
	switch m.Content.Type {

	case interfaces.ControlMessageTypeCreateRoom:
		err = h.createRoomAction(m)

	case interfaces.ControlMessageTypeJoinRoom:
		err = h.joinRoomAction(m)

	case interfaces.ControlMessageTypeTerminateRoom:
		err = h.terminateRoomAction(m)
	}

	if err != nil {
		response := messages.New(interfaces.TypeMessageError, err.Error())
		h.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
	}
}

func (h *Hub) passMessageToRoom(m *interfaces.IncomingMessage) {
	if m.Author.Room() == nil {
		response := messages.New(interfaces.TypeMessageError, NotInARoom)
		h.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
		return
	}

	defer func() {
		if rc := recover(); rc != nil {
			fmt.Printf("Panic in room '%s': %s\n", m.Author.Room().ID(), rc)
			debug.PrintStack()
			wg.Add(1)
			go h.destroyRoomConcurrently(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedGamePanicked)
		}
	}()

	m.Author.Room().Parse(m)
}

// Removes a client from the hub and also from a room if it's in one
func (h *Hub) removeClient(c interfaces.Client) {
	mutex.Lock()
	defer mutex.Unlock()
	for i := range h.clients {
		if h.clients[i] == c {
			if c.Room() != nil {
				r := c.Room()
				r.RemoveClient(c)
			}
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			if h.configuration.Debug {
				log.Printf("Client removed from hub, number of clients left: %d\n", len(h.clients))
			}
			c.Close()
			return
		}
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients() int {
	return len(h.clients)
}

func (h *Hub) createUpdatedRoomListMessage() interface{} {
	return messages.New(interfaces.TypeMessageRoomsList, h.getWaitingRoomsIds())
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

func (h *Hub) sendMessage(c interfaces.Client, message interface{}, typeName string, optArgs ...interface{}) {
	defer wg.Done()

	encoded := encodeMessage(message, typeName, optArgs)

	if h.configuration.Debug {
		log.Printf("Sending message %s to client '%s'\n", string(encoded[:]), c.Name())
	}

	select {
	case c.Incoming() <- encoded:
		return

	// We can't reach the client
	default:
		wg.Wait()
		h.removeClient(c)
		return
	}
}

func encodeMessage(message interface{}, typeName string, optArgs []interface{}) []byte {
	encodedContent, _ := json.Marshal(message)

	wrappedMessage := interfaces.OutgoingMessage{
		Type:    typeName,
		Content: encodedContent,
	}

	if len(optArgs) > 0 {
		wrappedMessage.SequenceNumber = optArgs[0].(int)
	}

	encoded, _ := json.Marshal(wrappedMessage)
	return encoded
}
