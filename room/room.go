// Package room contains the Room class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
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
	messages chan *interfaces.MessageFromClient

	// Bots unregistration requests
	unregister chan interfaces.Client

	// Maximum time this room instance will be kept alive
	timeout time.Duration

	// timer function that will close the room after X minutes
	timer *time.Timer
}

// New returns a new Room instance
func New(id string, b interfaces.Bridge, owner interfaces.Client, messages chan *interfaces.MessageFromClient, unregister chan interfaces.Client, cfg *config.Config) *Room {
	return &Room{
		id:         id,
		clients:    []interfaces.Client{},
		gameBridge: b,
		timeout:    cfg.Timeout,
		owner:      owner,
		messages:   messages,
		unregister: unregister,
	}
}

func (r *Room) ParseMessage(m *interfaces.MessageFromClient) (map[interfaces.Client][]byte, error) {
	if r.isControlMessage(m) {
		return r.parseControlMessage(m)
	}
	if r.gameBridge.IsGameOver() {
		return nil, errors.New(GameOver)
	}
	return r.passMessageToGame(m)
}

func (r *Room) isControlMessage(m *interfaces.MessageFromClient) bool {
	switch m.Content.Type {
	case
		interfaces.ControlMessageTypeAddBot,
		interfaces.ControlMessageTypeStartGame,
		interfaces.ControlMessageTypeKickPlayer,
		interfaces.ControlMessageTypePlayerQuits:
		return true
	}
	return false
}

func (r *Room) parseControlMessage(m *interfaces.MessageFromClient) (map[interfaces.Client][]byte, error) {
	response := map[interfaces.Client][]byte{}

	switch m.Content.Type {

	case interfaces.ControlMessageTypeStartGame:
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

	case interfaces.ControlMessageTypeAddBot:
		if m.Author != r.owner {
			return nil, errors.New(Forbidden)
		}
		var parsed interfaces.MessageAddBotParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			return r.addBot(parsed.BotLevel)
		}

	case interfaces.ControlMessageTypeKickPlayer:
		if m.Author != r.owner {
			return nil, errors.New(Forbidden)
		}
		var parsed interfaces.MessageKickPlayerParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			return r.kickClient(parsed.PlayerNumber)
		}

	case interfaces.ControlMessageTypePlayerQuits:
		return r.clientQuits(m.Author)

	}

	return response, nil
}

func (r *Room) passMessageToGame(m *interfaces.MessageFromClient) (map[interfaces.Client][]byte, error) {
	var err error
	var currentPlayer interfaces.Client
	response := map[interfaces.Client][]byte{}

	if currentPlayer, err = r.currentPlayerClient(); m.Author == currentPlayer && err == nil {
		err = r.gameBridge.Execute(m.Author.Name(), m.Content.Type, m.Content.Params)
		if err == nil {
			for n, cl := range r.clients {
				if cl != nil && cl.IsBot() && r.IsGameOver() {
					continue
				}
				if cl != nil {
					response[cl], _ = r.gameBridge.Status(n)
				}
			}
		}
	}
	return response, nil
}

func (r *Room) addBot(level string) (map[interfaces.Client][]byte, error) {
	var err error
	var c interfaces.Client
	response := map[interfaces.Client][]byte{}

	if c, err = r.gameBridge.AddBot(level, r); err == nil {
		c.SetName(fmt.Sprintf("Bot %d", len(r.clients)+1))
		if response, err = r.AddClient(c); err == nil {
			go c.WritePump()
			go c.ReadPump(r.messages, r.unregister)

			return response, nil
		}
	}
	return nil, err
}

func (r *Room) updatedPlayersList() []byte {
	msg := interfaces.MessageCurrentPlayers{
		Type:   "pls",
		Values: r.playerData(),
	}
	response, _ := json.Marshal(msg)
	return response
}

func (r *Room) playerData() []interfaces.MessagePlayer {
	players := []interfaces.MessagePlayer{}
	for _, c := range r.clients {
		if c != nil {
			players = append(
				players,
				interfaces.MessagePlayer{
					Name:  c.Name(),
					Owner: c.Room().Owner() == c,
				},
			)
		}
	}
	return players
}

func (r *Room) currentPlayerClient() (interfaces.Client, error) {
	number, err := r.gameBridge.CurrentPlayerNumber()
	return r.clients[number], err
}

func (r *Room) startGame() error {
	return r.gameBridge.StartGame()
}

func (r *Room) AddClient(c interfaces.Client) (map[interfaces.Client][]byte, error) {
	response := map[interfaces.Client][]byte{}

	if err := r.gameBridge.AddPlayer(c.Name()); err != nil {
		return nil, err
	}
	r.clients = append(r.clients, c)

	if len(r.clients) == 1 {
		r.owner = c
	}
	c.SetRoom(r)
	for _, cl := range r.clients {
		response[cl] = r.updatedPlayersList()
	}

	return response, nil
}

func (r *Room) kickClient(number int) (map[interfaces.Client][]byte, error) {
	response := map[interfaces.Client][]byte{}

	if number < 0 || number >= len(r.clients) {
		return nil, errors.New(InexistentClient)
	}
	cl := r.clients[number]
	if cl == r.owner {
		return nil, errors.New(OwnerNotRemovable)
	}
	cl.SetRoom(nil)
	response = r.RemoveClient(r.clients[number])
	msg := interfaces.MessageRoomDestroyed{
		Type:   "out",
		Reason: "kck",
	}
	encodedMsg, _ := json.Marshal(msg)
	response[cl] = encodedMsg
	return response, nil
}

func (r *Room) clientQuits(cl interfaces.Client) (map[interfaces.Client][]byte, error) {
	response := map[interfaces.Client][]byte{}

	response = r.RemoveClient(cl)
	msg := interfaces.MessageRoomDestroyed{
		Type:   "out",
		Reason: "qui",
	}
	encodedMsg, _ := json.Marshal(msg)
	response[cl] = encodedMsg
	return response, nil
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (r *Room) RemoveClient(c interfaces.Client) map[interfaces.Client][]byte {
	response := map[interfaces.Client][]byte{}

	for i := range r.clients {
		if r.clients[i] == c {
			r.clients[i].SetRoom(nil)
			if r.gameBridge.GameStarted() {
				response = r.deactivatePlayer(i)
			} else {
				response = r.removePlayer(i)
			}
			break
		}
	}
	return response
}

// deactivatePlayerResponse deactivates a player from a game setting it as nil,
// and returns an updated game status to all the players as a response
func (r *Room) deactivatePlayer(playerNumber int) map[interfaces.Client][]byte {
	response := map[interfaces.Client][]byte{}
	r.clients[playerNumber] = nil
	r.gameBridge.DeactivatePlayer(playerNumber)
	for i, cl := range r.clients {
		if cl == nil || cl.IsBot() {
			continue
		}
		response[cl], _ = r.gameBridge.Status(i)
	}
	return response
}

// deactivatePlayerResponse removes a player from a room,
// and returns an updated players list to all the clients as a response
func (r *Room) removePlayer(playerNumber int) map[interfaces.Client][]byte {
	response := map[interfaces.Client][]byte{}
	r.clients = append(r.clients[:playerNumber], r.clients[playerNumber+1:]...)
	r.gameBridge.RemovePlayer(playerNumber)
	for _, cl := range r.clients {
		response[cl] = r.updatedPlayersList()
	}
	return response
}

func (r *Room) GameStarted() bool {
	return r.gameBridge.GameStarted()
}

func (r *Room) IsGameOver() bool {
	return r.gameBridge.IsGameOver()
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) Owner() interfaces.Client {
	return r.owner
}

func (r *Room) Clients() []interfaces.Client {
	return r.clients
}

func (r *Room) SetTimer(t *time.Timer) {
	r.timer = t
}

func (r *Room) Timer() *time.Timer {
	return r.timer
}
