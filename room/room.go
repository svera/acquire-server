// Package room contains the Room class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package room

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
	GameOver          = "game_over"
)

// Room is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Bridge
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Room struct {
	id string

	// Registered clients
	clients []interfaces.Client

	owner interfaces.Client

	gameBridge interfaces.Bridge

	// Bots inbound messages
	messages chan *interfaces.ClientMessage

	// Bots unregistration requests
	unregister chan interfaces.Client

	// Maximum time this room instance will be kept alive
	timeout time.Duration

	wasClosedByTimeout bool
}

// New returns a new Room instance
func New(id string, b interfaces.Bridge, owner interfaces.Client, messages chan *interfaces.ClientMessage, unregister chan interfaces.Client, cfg *config.Config) *Room {
	return &Room{
		id:                 id,
		clients:            []interfaces.Client{},
		gameBridge:         b,
		timeout:            cfg.Timeout,
		wasClosedByTimeout: false,
		owner:              owner,
		messages:           messages,
		unregister:         unregister,
	}
}

func (r *Room) ParseMessage(m *interfaces.ClientMessage) (map[interfaces.Client][]byte, error) {
	var err error
	var currentPlayer interfaces.Client
	response := map[interfaces.Client][]byte{}

	if r.gameBridge.IsGameOver() {
		return nil, errors.New(GameOver)
	}

	if r.isControlMessage(m) {
		return r.parseControlMessage(m)
	} else {
		if currentPlayer, err = r.currentPlayerClient(); m.Author == currentPlayer && err == nil {
			err = r.gameBridge.Execute(m.Content.Type, m.Content.Params)
			if err == nil {
				for n, cl := range r.clients {
					if cl.IsBot() && r.IsGameOver() {
						continue
					}
					response[cl], _ = r.gameBridge.Status(n)
				}
				return response, nil
			}
		}
	}
	if err != nil {
		log.Println(err)
	}
	return nil, err
}

func (r *Room) isControlMessage(m *interfaces.ClientMessage) bool {
	switch m.Content.Type {
	case
		client.ControlMessageTypeJoinPlayer,
		client.ControlMessageTypeAddBot,
		client.ControlMessageTypeStartGame,
		client.ControlMessageTypeKickPlayer,
		client.ControlMessageTypePlayerQuits:
		return true
	}
	return false
}

func (r *Room) parseControlMessage(m *interfaces.ClientMessage) (map[interfaces.Client][]byte, error) {
	response := map[interfaces.Client][]byte{}

	switch m.Content.Type {

	case client.ControlMessageTypeStartGame:
		if m.Author != r.owner {
			return nil, errors.New(Forbidden)
		}
		if err := r.startGame(); err != nil {
			return nil, err
		}
		for n, cl := range r.clients {
			st, _ := r.gameBridge.Status(n)
			response[cl] = st
		}

	case client.ControlMessageTypeJoinPlayer:
		if err := r.addClient(m.Author); err != nil {
			return nil, err
		}
		log.Printf("Cliente añadido, Numero de clientes: %d\n", len(r.clients))

		for _, cl := range r.clients {
			response[cl] = r.updatedPlayersList()
		}

	case client.ControlMessageTypeAddBot:
		if m.Author != r.owner {
			return nil, errors.New(Forbidden)
		}
		var parsed client.AddBotMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if cl, err := r.addBot(parsed.BotName); err != nil {
				return nil, err
			} else {
				go cl.WritePump()
				go cl.ReadPump(r.messages, r.unregister)
				for _, cl := range r.clients {
					response[cl] = r.updatedPlayersList()
				}
			}
		}

	case client.ControlMessageTypeKickPlayer:
		if m.Author != r.owner {
			return nil, errors.New(Forbidden)
		}
		var parsed client.KickPlayerMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if err := r.kickClient(parsed.PlayerNumber); err != nil {
				return nil, err
			} else {
				for _, cl := range r.clients {
					response[cl] = r.updatedPlayersList()
				}
			}
		}
		/*
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
	return response, nil
}

func (r *Room) addBot(level string) (interfaces.Client, error) {
	var err error
	var c interfaces.Client
	if c, err = r.gameBridge.AddBot(level, r); err == nil {
		if err = r.addClient(c); err == nil {
			c.SetName(fmt.Sprintf("Bot %d", r.NumberClients()+1))
			return c, nil
		}
	}
	return nil, err
}

func (r *Room) updatedPlayersList() []byte {
	msg := currentPlayersMessage{
		Type:   "pls",
		Values: r.clientNames(),
	}
	response, _ := json.Marshal(msg)
	return response
}

func (r *Room) clientNames() []string {
	names := []string{}
	for _, c := range r.clients {
		if c != nil {
			names = append(names, c.Name())
		}
	}
	return names
}

func (r *Room) currentPlayerClient() (interfaces.Client, error) {
	number, err := r.gameBridge.CurrentPlayerNumber()
	return r.clients[number], err
}

func (r *Room) startGame() error {
	return r.gameBridge.StartGame()
}

func (r *Room) addClient(c interfaces.Client) error {
	if err := r.gameBridge.AddPlayer(c.Name()); err != nil {
		return err
	}
	r.clients = append(r.clients, c)

	if len(r.clients) == 1 {
		r.owner = c
	}
	c.SetRoom(r)
	return nil
}

func (r *Room) kickClient(number int) error {
	if number < 0 || number > len(r.clients) {
		return errors.New(InexistentClient)
	}
	if r.clients[number] == r.owner {
		return errors.New(OwnerNotRemovable)
	}
	r.clients[number].Close(interfaces.PlayerKicked)
	r.RemoveClient(r.clients[number])
	return nil
}

func (r *Room) quitClient(client interfaces.Client) error {
	if client == r.owner {
		return errors.New(OwnerNotRemovable)
	}
	client.Close(interfaces.PlayerQuit)
	r.RemoveClient(client)
	return nil
}

func (r *Room) terminateGame(client interfaces.Client) error {
	if client != r.owner {
		return errors.New(Forbidden)
	}
	r.closeRoom()
	return nil
}

func (r *Room) closeRoom() {
	for _, cl := range r.clients {
		if cl != nil {
			if r.wasClosedByTimeout {
				cl.Close(interfaces.HubTimeout)
			} else if r.gameBridge.IsGameOver() {
				cl.Close(interfaces.EndOk)
			} else {
				cl.Close(interfaces.HubDestroyed)
			}
		}
	}
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (r *Room) RemoveClient(c interfaces.Client) map[interfaces.Client][]byte {
	response := map[interfaces.Client][]byte{}

	for i := range r.clients {
		if r.clients[i] == c {
			if r.gameBridge.GameStarted() {
				r.clients[i] = nil
				r.gameBridge.DeactivatePlayer(i)
			} else {
				r.clients = append(r.clients[:i], r.clients[i+1:]...)
				r.gameBridge.RemovePlayer(i)
			}
			for _, cl := range r.clients {
				if cl != nil {
					response[cl] = r.updatedPlayersList()
				}
			}

			log.Printf("Cliente eliminado de la habitacion, Numero de clientes: %d\n", len(r.clients))
			break
		}
	}
	return response
}

func (r *Room) GameStarted() bool {
	return r.gameBridge.GameStarted()
}

// NumberClients returns the number of connected clients in this room
func (r *Room) NumberClients() int {
	return len(r.clients)
}

func (r *Room) IsGameOver() bool {
	return r.gameBridge.IsGameOver()
}

func (r *Room) ID() string {
	return r.id
}
