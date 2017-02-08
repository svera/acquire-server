package hub

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
	"github.com/svera/sackson-server/room"
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

	callbacks map[string]func(...interface{})
}

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rn = rand.New(source)
}

// New returns a new Hub instance
func New(cfg *config.Config) *Hub {
	h := &Hub{
		Messages:      make(chan *interfaces.IncomingMessage),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
		callbacks:     make(map[string]func(...interface{})),
	}

	h.registerCallbacks()

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
			h.callbacks["messageCreated"]([]interfaces.Client{cl}, h.createUpdatedRoomListMessage())
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
		h.callbacks["messageCreated"]([]interfaces.Client{m.Author}, response)
	}
}

func (h *Hub) passMessageToRoom(m *interfaces.IncomingMessage) {
	if m.Author.Room() == nil {
		response := messages.New(interfaces.TypeMessageError, NotInARoom)
		h.callbacks["messageCreated"]([]interfaces.Client{m.Author}, response)
		return
	}

	m.Author.Room().Parse(m)
}

func (h *Hub) sendMessage(c interfaces.Client, message []byte) {
	log.Printf("Sending message %s to client '%s'\n", string(message[:]), c.Name())
	defer wg.Done()

	select {
	case c.Incoming() <- message:
		return

	// We can't reach the client
	default:
		wg.Wait()
		h.removeClient(c)
		return
	}
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
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
			//close(c.Incoming())
			c.Close()
			return
		}
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients() int {
	return len(h.clients)
}

func (h *Hub) createUpdatedRoomListMessage() []byte {
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

func (h *Hub) registerCallbacks() {
	h.callbacks[room.GameStarted] = func(args ...interface{}) {
		message := h.createUpdatedRoomListMessage()

		wg.Add(len(h.clients))
		for _, cl := range h.clients {
			go h.sendMessage(cl, message)
		}
	}

	h.callbacks["messageCreated"] = func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)
		message := args[1].([]byte)

		wg.Add(len(clients))
		//mutex.Lock()
		for _, cl := range clients {
			go h.sendMessage(cl, message)
		}
		//mutex.Unlock()
	}

	h.callbacks[room.ClientOut] = func(args ...interface{}) {
		r := args[0].(interfaces.Room)

		if len(r.HumanClients()) == 0 {
			wg.Add(1)
			go h.destroyRoomWithoutHumansAction(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
	}
}
