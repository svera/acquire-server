// Package room contains the Room class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	observable "github.com/GianlucaGuarini/go-observable"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
)

// Error messages returned from Room
const (
	InexistentClient  = "inexistent_client"
	OwnerNotRemovable = "owner_not_removable"
	Forbidden         = "forbidden"
	GameOver          = "game_over"
)

// Events triggered from Room, always in a past tense
const (
	GameStarted = "gameStarted"
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

	observer *observable.Observable
}

// New returns a new Room instance
func New(
	id string, b interfaces.Bridge,
	owner interfaces.Client,
	messages chan *interfaces.MessageFromClient,
	unregister chan interfaces.Client,
	cfg *config.Config,
	observer *observable.Observable,
) *Room {
	return &Room{
		id:         id,
		clients:    []interfaces.Client{},
		gameBridge: b,
		timeout:    cfg.Timeout,
		owner:      owner,
		messages:   messages,
		unregister: unregister,
		observer:   observer,
	}
}

// ParseMessage gets an incoming message from a client and parses it, executing
// its desired action in the room or passing it to the room's game bridge
func (r *Room) ParseMessage(m *interfaces.MessageFromClient) {
	if r.isControlMessage(m) {
		r.parseControlMessage(m)
	} else if r.gameBridge.IsGameOver() {
		response := newMessage(interfaces.TypeMessageError, GameOver)
		r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
	} else {
		r.passMessageToGame(m)
	}
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

func (r *Room) parseControlMessage(m *interfaces.MessageFromClient) {
	var err error
	switch m.Content.Type {

	case interfaces.ControlMessageTypeStartGame:
		err = r.startGameAction(m)

	case interfaces.ControlMessageTypeAddBot:
		err = r.addBotAction(m)

	case interfaces.ControlMessageTypeKickPlayer:
		err = r.kickPlayerAction(m)

	case interfaces.ControlMessageTypePlayerQuits:
		err = r.clientQuits(m.Author)
	}
	if err != nil {
		response := newMessage(interfaces.TypeMessageError, err.Error())
		r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
	}
}

func (r *Room) passMessageToGame(m *interfaces.MessageFromClient) {
	var err error
	var currentPlayer interfaces.Client

	if currentPlayer, err = r.currentPlayerClient(); m.Author == currentPlayer && err == nil {
		err = r.gameBridge.Execute(m.Author.Name(), m.Content.Type, m.Content.Params)
		if err == nil {
			for n, cl := range r.clients {
				if cl != nil && cl.IsBot() && r.IsGameOver() {
					continue
				}
				if cl != nil {
					st, _ := r.gameBridge.Status(n)
					r.observer.Trigger("messageCreated", []interfaces.Client{cl}, st)
				}
			}
		} else {
			response := newMessage(interfaces.TypeMessageError, err.Error())
			r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
		}
	}
}

func (r *Room) startGameAction(m *interfaces.MessageFromClient) error {
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	if err := r.startGame(); err != nil {
		return err
	}
	for n, cl := range r.clients {
		st, _ := r.gameBridge.Status(n)
		r.observer.Trigger("messageCreated", []interfaces.Client{cl}, st)
	}
	r.observer.Trigger(GameStarted)
	return nil
}

func (r *Room) addBotAction(m *interfaces.MessageFromClient) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed interfaces.MessageAddBotParams
	if err = json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if err = r.addBot(parsed.BotLevel); err != nil {
			response := newMessage(interfaces.TypeMessageError, err.Error())
			r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
		}
	}
	return err
}

func (r *Room) kickPlayerAction(m *interfaces.MessageFromClient) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed interfaces.MessageKickPlayerParams
	if err = json.Unmarshal(m.Content.Params, &parsed); err == nil {
		return r.kickClient(parsed.PlayerNumber)
	}
	return err
}

func (r *Room) addBot(level string) error {
	var err error
	var c interfaces.Client

	if c, err = r.gameBridge.AddBot(level, r); err == nil {
		c.SetName(fmt.Sprintf("Bot %d", len(r.clients)+1))
		if err = r.addClient(c); err == nil {
			go c.WritePump()
			go c.ReadPump(r.messages, r.unregister)
		}
	}
	return err
}

func (r *Room) playersData() []interfaces.MessagePlayer {
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

// AddHuman adds a new client to the room
func (r *Room) AddHuman(c interfaces.Client) error {
	return r.addClient(c)
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
	response := newMessage(interfaces.TypeMessageCurrentPlayers, r.playersData())
	r.observer.Trigger("messageCreated", r.clients, response)
	return nil
}

func (r *Room) kickClient(number int) error {
	if number < 0 || number >= len(r.clients) {
		return errors.New(InexistentClient)
	}
	cl := r.clients[number]
	if cl == r.owner {
		return errors.New(OwnerNotRemovable)
	}
	cl.SetRoom(nil)
	r.RemoveClient(r.clients[number])
	response := newMessage(interfaces.TypeMessageRoomDestroyed, "kck")
	r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response)
	return nil
}

func (r *Room) clientQuits(cl interfaces.Client) error {
	r.RemoveClient(cl)
	response := newMessage(interfaces.TypeMessageRoomDestroyed, "qui")
	r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response)
	return nil
}

// RemoveClient Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (r *Room) RemoveClient(c interfaces.Client) {
	for i := range r.clients {
		if r.clients[i] == c {
			r.clients[i].SetRoom(nil)
			if r.gameBridge.GameStarted() {
				r.deactivatePlayer(i)
			} else {
				r.removePlayer(i)
			}
			break
		}
	}
}

// deactivatePlayer deactivates a player from a game setting it as nil,
// and returns an updated game status to all the players as a response
func (r *Room) deactivatePlayer(playerNumber int) {
	r.clients[playerNumber] = nil
	r.gameBridge.DeactivatePlayer(playerNumber)
	for i, cl := range r.clients {
		if cl == nil || cl.IsBot() {
			continue
		}
		st, _ := r.gameBridge.Status(i)
		r.observer.Trigger("messageCreated", []interfaces.Client{cl}, st)
	}
}

// removePlayer removes a player from a room,
// and returns an updated players list to all the clients as a response
func (r *Room) removePlayer(playerNumber int) {
	r.clients = append(r.clients[:playerNumber], r.clients[playerNumber+1:]...)
	r.gameBridge.RemovePlayer(playerNumber)
	response := newMessage(interfaces.TypeMessageCurrentPlayers, r.playersData())
	for _, cl := range r.clients {
		r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response)
	}
}

// GameStarted returns true if the room's game has started, false otherwise
func (r *Room) GameStarted() bool {
	return r.gameBridge.GameStarted()
}

// IsGameOver returns true if the room's game has ended, false otherwise
func (r *Room) IsGameOver() bool {
	return r.gameBridge.IsGameOver()
}

// ID returns the room's ID
func (r *Room) ID() string {
	return r.id
}

// Owner returns the room's owner
func (r *Room) Owner() interfaces.Client {
	return r.owner
}

// Clients returns the room's connected clients
func (r *Room) Clients() []interfaces.Client {
	return r.clients
}

// HumanClients returns room's connected human clients
func (r *Room) HumanClients() []interfaces.Client {
	human := []interfaces.Client{}
	for _, c := range r.clients {
		if c != nil && !c.IsBot() {
			human = append(human, c)
		}
	}
	return human
}

// SetTimer sets the room's timer, that manages when to close a room
func (r *Room) SetTimer(t *time.Timer) {
	r.timer = t
}

// Timer returns the room's timer
func (r *Room) Timer() *time.Timer {
	return r.timer
}
