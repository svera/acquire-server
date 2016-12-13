package room

import (
	"log"
	"sync"
	"time"

	emitable "github.com/olebedev/emitter"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

// Error messages returned from Room
const (
	InexistentClient  = "inexistent_client"
	OwnerNotRemovable = "owner_not_removable"
	Forbidden         = "forbidden"
	GameOver          = "game_over"
)

// Events Emited from Room, always in a past tense
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
	messages chan *interfaces.IncomingMessage

	// Bots unregistration requests
	unregister chan interfaces.Client

	// Maximum time this room instance will be kept alive
	timeout time.Duration

	// timer function that will close the room after X minutes
	timer *time.Timer

	emitter *emitable.Emitter

	clientInTurn interfaces.Client

	playerTimeOut time.Duration

	mu sync.Mutex
}

// New returns a new Room instance
func New(
	id string, b interfaces.Bridge,
	owner interfaces.Client,
	messages chan *interfaces.IncomingMessage,
	unregister chan interfaces.Client,
	cfg *config.Config,
	emitter *emitable.Emitter,
	roomParams map[string]interface{},
) *Room {
	return &Room{
		id:            id,
		clients:       []interfaces.Client{},
		gameBridge:    b,
		timeout:       cfg.Timeout,
		owner:         owner,
		messages:      messages,
		unregister:    unregister,
		emitter:       emitter,
		clientInTurn:  nil,
		playerTimeOut: roomParams["playerTimeout"].(time.Duration),
	}
}

// Parse gets an incoming message from a client and parses it, executing
// its desired action in the room or passing it to the room's game bridge
func (r *Room) Parse(m *interfaces.IncomingMessage) {
	if r.isControlMessage(m) {
		r.parseControlMessage(m)
	} else if r.gameBridge.IsGameOver() {
		response := messages.New(interfaces.TypeMessageError, GameOver)
		go r.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
	} else {
		r.passMessageToGame(m)
	}
}

func (r *Room) isControlMessage(m *interfaces.IncomingMessage) bool {
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

func (r *Room) parseControlMessage(m *interfaces.IncomingMessage) {
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
		response := messages.New(interfaces.TypeMessageError, err.Error())
		go r.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
	}
}

func (r *Room) passMessageToGame(m *interfaces.IncomingMessage) {
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
					response, _ := r.gameBridge.Status(n)
					go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, response)
				}
			}
			currentPlayerClient, _ := r.currentPlayerClient()
			if r.clientInTurn != currentPlayerClient {
				r.changePlayerSetTimer()
			}
		} else {
			response := messages.New(interfaces.TypeMessageError, err.Error())
			go r.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
		}
	}
}

func (r *Room) changePlayerSetTimer() {
	if r.clientInTurn != nil {
		r.clientInTurn.StopTimer()
	}
	r.clientInTurn, _ = r.currentPlayerClient()
	if !r.clientInTurn.IsBot() && r.playerTimeOut > 0 {
		r.clientInTurn.StartTimer(time.Second * r.playerTimeOut)
	}
}

func (r *Room) playersData() []interfaces.PlayerData {
	players := []interfaces.PlayerData{}
	for _, c := range r.clients {
		if c != nil {
			players = append(
				players,
				interfaces.PlayerData{
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

// AddHuman adds a new client to the room
func (r *Room) AddHuman(c interfaces.Client) error {
	var err error
	if err = r.addClient(c); err == nil {
		if r.playerTimeOut > 0 {
			c.SetTimer(time.AfterFunc(time.Second*r.playerTimeOut, func() {
				log.Printf("client %s timed out", c.Name())
				r.timeoutPlayer(c)
			}))
		}
		return nil
	}
	return err
}

func (r *Room) timeoutPlayer(cl interfaces.Client) {
	response := messages.New(interfaces.TypeMessageClientOut, interfaces.ReasonPlayerTimedOut)

	go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, response)
	r.RemoveClient(cl)
}

func (r *Room) addClient(c interfaces.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.gameBridge.AddPlayer(c.Name()); err != nil {
		return err
	}
	r.clients = append(r.clients, c)

	if len(r.clients) == 1 {
		r.owner = c
	}
	c.SetRoom(r)
	response := messages.New(interfaces.TypeMessageCurrentPlayers, r.playersData())
	go r.emitter.Emit("messageCreated", r.clients, response)
	return nil
}

// RemoveClient removes / sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (r *Room) RemoveClient(c interfaces.Client) {
	for i := range r.clients {
		if r.clients[i] == c {
			r.clients[i].SetRoom(nil)
			c.StopTimer()
			if r.gameBridge.GameStarted() {
				r.deactivatePlayer(i)
			} else {
				r.removePlayer(i)
			}
			go r.emitter.Emit("clientOut", r)
			break
		}
	}
}

// deactivatePlayer deactivates a player from a game setting it as nil,
// and returns an updated game status to all the players as a response
func (r *Room) deactivatePlayer(playerNumber int) {
	r.clients[playerNumber] = nil
	r.gameBridge.DeactivatePlayer(playerNumber)
	if !r.gameBridge.IsGameOver() {
		currentPlayerClient, _ := r.currentPlayerClient()
		if r.clientInTurn != currentPlayerClient {
			r.changePlayerSetTimer()
		}
	}

	for i, cl := range r.clients {
		if cl == nil || cl.IsBot() {
			continue
		}
		st, _ := r.gameBridge.Status(i)
		go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, st)
	}
}

// removePlayer removes a player from a room,
// and returns an updated players list to all the clients as a response
func (r *Room) removePlayer(playerNumber int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients = append(r.clients[:playerNumber], r.clients[playerNumber+1:]...)
	r.gameBridge.RemovePlayer(playerNumber)
	response := messages.New(interfaces.TypeMessageCurrentPlayers, r.playersData())
	for _, cl := range r.clients {
		go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, response)
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
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.clients
}

// HumanClients returns room's connected human clients
func (r *Room) HumanClients() []interfaces.Client {
	r.mu.Lock()
	defer r.mu.Unlock()
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
