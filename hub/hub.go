package hub

import (
	"log"
	"math/rand"
	"sync"
	"time"

	emitable "github.com/olebedev/emitter"
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
	wg      sync.WaitGroup
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

	emitter *emitable.Emitter

	mu sync.Mutex
}

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rn = rand.New(source)
}

// New returns a new Hub instance
func New(cfg *config.Config, emitter *emitable.Emitter) *Hub {
	h := &Hub{
		Messages:      make(chan *interfaces.IncomingMessage),
		Register:      make(chan interfaces.Client),
		Unregister:    make(chan interfaces.Client),
		clients:       []interfaces.Client{},
		rooms:         make(map[string]interfaces.Room),
		configuration: cfg,
		emitter:       emitter,
	}

	h.registerCallbacks()

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
			h.mu.Lock()
			h.clients = append(h.clients, c)
			h.mu.Unlock()
			go h.emitter.Emit("messageCreated", h.clients, h.createUpdatedRoomListMessage())
			if h.configuration.Debug {
				log.Printf("Client added to hub, number of connected clients: %d\n", len(h.clients))
			}

		case c := <-h.Unregister:
			for _, val := range h.clients {
				if val == c {
					wg.Wait()
					h.removeClient(c)
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
	switch m.Content.Type {

	case interfaces.ControlMessageTypeCreateRoom:
		h.createRoomAction(m)

	case interfaces.ControlMessageTypeJoinRoom:
		h.joinRoomAction(m)

	case interfaces.ControlMessageTypeTerminateRoom:
		h.terminateRoomAction(m)
	}
}

func (h *Hub) passMessageToRoom(m *interfaces.IncomingMessage) {
	m.Author.Room().Parse(m)
}

func (h *Hub) sendMessage(c interfaces.Client, message []byte) {
	wg.Add(1)

	select {
	case c.Incoming() <- message:
		wg.Done()
		break

	// We can't reach the client
	default:
		wg.Wait()
		h.removeClient(c)
	}
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (h *Hub) removeClient(c interfaces.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
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
			break
		}
	}
}

// NumberClients returns the number of connected clients
func (h *Hub) NumberClients() int {
	return len(h.clients)
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
	h.emitter.On(room.GameStarted, func(event *emitable.Event) {
		message := h.createUpdatedRoomListMessage()

		for _, cl := range h.clients {
			h.sendMessage(cl, message)
		}
	})

	h.emitter.On("messageCreated", func(event *emitable.Event) {
		clients := event.Args[0].([]interfaces.Client)
		message := event.Args[1].([]byte)

		h.mu.Lock()
		for _, cl := range clients {
			h.sendMessage(cl, message)
		}
		h.mu.Unlock()
	})

	h.emitter.On("clientOut", func(event *emitable.Event) {
		r := event.Args[0].(interfaces.Room)

		if len(r.HumanClients()) == 0 {
			h.destroyRoom(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
	})
}
