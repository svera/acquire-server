// Package hub contains the Hub class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/config"
	"github.com/svera/tbg-server/interfaces"
)

const (
	InexistentClient  = "inexistent_client"
	OwnerNotRemovable = "owner_not_removable"
	Forbidden         = "forbidden"
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

	// Registration requests
	Register chan interfaces.Client

	// Unregistration requests
	Unregister chan interfaces.Client

	// Stops hub server
	stop chan struct{}

	gameBridge interfaces.Bridge

	// This callBack is executed when the hub stops running to remove it from memory
	selfDestructCallBack func()

	// Maximum time this hub instance will be kept alive
	timeout time.Duration

	wasClosedByTimeout bool
}

// New returns a new Hub instance
func New(b interfaces.Bridge, callBack func(), cfg *config.Config) *Hub {
	return &Hub{
		Messages:             make(chan *client.Message),
		Register:             make(chan interfaces.Client),
		Unregister:           make(chan interfaces.Client),
		stop:                 make(chan struct{}),
		clients:              []interfaces.Client{},
		gameBridge:           b,
		selfDestructCallBack: callBack,
		timeout:              cfg.Timeout,
		wasClosedByTimeout:   false,
	}
}

// Run listens for messages coming from several channels and acts accordingly
func (h *Hub) Run() {
	defer h.selfDestructCallBack()

	time.AfterFunc(time.Minute*h.timeout, func() {
		h.wasClosedByTimeout = true
		h.stopHub()
	})

	for {
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
					if c.Owner() {
						h.stopHub()
						return
					}
				}
			}
			break

		case <-h.stop:
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
	} else if !h.gameBridge.IsGameOver() {
		h.parseGameMessage(m)
	}
}

func (h *Hub) isControlMessage(m *client.Message) bool {
	switch m.Content.Type {
	case
		client.ControlMessageTypeAddBot,
		client.ControlMessageTypeStartGame,
		client.ControlMessageTypeKickPlayer,
		client.ControlMessageTypeTerminateGame,
		client.ControlMessageTypePlayerQuits:
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
	case client.ControlMessageTypePlayerQuits:
		if err := h.quitClient(m.Author); err != nil {
			h.sendErrorMessage(err, m.Author)
		}
	case client.ControlMessageTypeTerminateGame:
		if err := h.terminateGame(m.Author); err != nil {
			h.sendErrorMessage(err, m.Author)
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
		if c != nil {
			names = append(names, c.Name())
		}
	}
	return names
}

func (h *Hub) broadcastUpdate() {
	for n, c := range h.clients {
		if c != nil {
			if c.IsBot() && h.gameBridge.IsGameOver() {
				continue
			}
			response, _ := h.gameBridge.Status(n)
			h.sendMessage(c, response)
		}
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

func (h *Hub) addClient(c interfaces.Client) error {
	name := fmt.Sprintf("Player %d", h.NumberClients()+1)
	c.SetName(name)

	if err := h.gameBridge.AddPlayer(name); err != nil {
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
	h.clients[number].Close(interfaces.PlayerKicked)
	h.removeClient(h.clients[number])
	return nil
}

func (h *Hub) quitClient(client interfaces.Client) error {
	if client.Owner() {
		return errors.New(OwnerNotRemovable)
	}
	client.Close(interfaces.PlayerQuit)
	h.removeClient(client)
	return nil
}

func (h *Hub) terminateGame(client interfaces.Client) error {
	if !client.Owner() {
		return errors.New(Forbidden)
	}
	h.stopHub()
	return nil
}

func (h *Hub) stopHub() {
	for _, cl := range h.clients {
		if cl != nil {
			if h.wasClosedByTimeout {
				cl.Close(interfaces.HubTimeout)
			} else if h.gameBridge.IsGameOver() {
				cl.Close(interfaces.EndOk)
			} else {
				cl.Close(interfaces.HubDestroyed)
			}
		}
	}

	close(h.stop)
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (h *Hub) removeClient(c interfaces.Client) {
	for i := range h.clients {
		if h.clients[i] == c {
			if h.gameBridge.GameStarted() {
				h.clients[i] = nil
				h.gameBridge.DeactivatePlayer(i)
				h.broadcastUpdate()
			} else {
				h.clients = append(h.clients[:i], h.clients[i+1:]...)
				h.gameBridge.RemovePlayer(i)
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
