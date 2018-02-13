package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/svera/sackson-server/internal/config"
	"github.com/svera/sackson-server/internal/events"
	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
)

var (
	mutex sync.RWMutex
	rn    *rand.Rand
	wg    sync.WaitGroup
)

// Hub is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Driver
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Hub struct {
	// Registered clients
	clients map[string][]interfaces.Client

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
		clients:       map[string][]interfaces.Client{},
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
			h.clients[cl.Game()] = append(h.clients[cl.Game()], cl)
			mutex.Unlock()
			cl.SetName(fmt.Sprintf("Player %d", h.NumberClients(cl.Game())))
			h.observer.Trigger(events.ClientRegistered{Client: cl})

			if h.configuration.Debug {
				log.Printf("Client added to hub using game '%s', number of connected clients: %d\n", cl.Game(), len(h.clients))
			}

		case cl := <-h.Unregister:
			for _, val := range h.clients[cl.Game()] {
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
	switch m.Type {
	case
		messages.TypeCreateRoom,
		messages.TypeJoinRoom,
		messages.TypeTerminateRoom:
		return true
	}
	return false
}

func (h *Hub) parseControlMessage(m *interfaces.IncomingMessage) {
	var err error
	switch m.Type {

	case messages.TypeCreateRoom:
		err = h.createRoomAction(m)

	case messages.TypeJoinRoom:
		err = h.joinRoomAction(m)

	case messages.TypeTerminateRoom:
		err = h.terminateRoomAction(m)
	}

	if err != nil {
		h.observer.Trigger(events.Error{Client: m.Author, ErrorText: err.Error()})
	}
}

func (h *Hub) passMessageToRoom(m *interfaces.IncomingMessage) {
	if m.Author.Room() == nil {
		h.observer.Trigger(events.Error{Client: m.Author, ErrorText: NotInARoom})
		return
	}

	defer func() {
		if rc := recover(); rc != nil {
			fmt.Printf("Panic in room '%s': %s\n", m.Author.Room().ID(), rc)
			debug.PrintStack()
			go h.destroyRoom(m.Author.Room().ID(), messages.ReasonRoomDestroyedGamePanicked)
		}
	}()
	m.Author.Room().Parse(m)
}

// Removes a client from the hub and also from a room if it's in one
func (h *Hub) removeClient(cl interfaces.Client) {
	for i := range h.clients[cl.Game()] {
		if h.clients[cl.Game()][i] == cl {
			mutex.Lock()
			h.clients[cl.Game()] = append(h.clients[cl.Game()][:i], h.clients[cl.Game()][i+1:]...)
			mutex.Unlock()
			h.observer.Trigger(events.ClientUnregistered{Client: cl})
			if h.configuration.Debug {
				log.Printf("Client removed from hub, number of clients left: %d\n", len(h.clients[cl.Game()]))
			}
			cl.Close()
			return
		}
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients(game string) int {
	return len(h.clients[game])
}

func (h *Hub) createUpdatedRoomListMessage() interface{} {
	return messages.RoomsList{
		Values: h.getWaitingRoomsIds(),
	}
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
