package room

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

// Events emitted from Room, always in a past tense
const (
	GameStarted       = "gameStarted"
	ClientOut         = "clientOut"
	GameStatusUpdated = "gameStatusUpdated"
)

var (
	mutex sync.RWMutex
)

// Room is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Bridge
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Room struct {
	id string

	// Registered clients
	clients map[int]interfaces.Client

	owner interfaces.Client

	gameBridge interfaces.Bridge

	// Bots inbound messages
	messages chan *interfaces.IncomingMessage

	// Bots unregistration requests
	unregister chan interfaces.Client

	configuration *config.Config

	// timer function that will close the room after X minutes
	timer *time.Timer

	observer interfaces.Observer

	clientsInTurn []interfaces.Client

	playerTimeOut time.Duration

	clientCounter int

	updateSequenceNumber int
}

// New returns a new Room instance
func New(
	id string, b interfaces.Bridge,
	owner interfaces.Client,
	messages chan *interfaces.IncomingMessage,
	unregister chan interfaces.Client,
	cfg *config.Config,
	ob interfaces.Observer,
) *Room {
	return &Room{
		id:                   id,
		clients:              map[int]interfaces.Client{},
		gameBridge:           b,
		owner:                owner,
		messages:             messages,
		unregister:           unregister,
		observer:             ob,
		clientsInTurn:        nil,
		configuration:        cfg,
		clientCounter:        0,
		updateSequenceNumber: 0,
	}
}

// Parse gets an incoming message from a client and parses it, executing
// its desired action in the room or passing it to the room's game bridge
func (r *Room) Parse(m *interfaces.IncomingMessage) {
	if r.isControlMessage(m) {
		r.parseControlMessage(m)
	} else if r.gameBridge.IsGameOver() {
		response := messages.New(interfaces.TypeMessageError, GameOver)
		r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
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
		interfaces.ControlMessageTypePlayerQuits,
		interfaces.ControlMessageTypeSetClientData:
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

	case interfaces.ControlMessageTypeSetClientData:
		err = r.setClientDataAction(m)
	}

	if err != nil {
		response := messages.New(interfaces.TypeMessageError, err.Error())
		r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
	}
}

func (r *Room) passMessageToGame(m *interfaces.IncomingMessage) {
	var err error
	var st interface{}

	if r.messageAuthorIsInTurn(m) {
		if err = r.gameBridge.Execute(m.Author.Name(), m.Content.Type, m.Content.Params); err == nil {
			r.updateSequenceNumber++
			for n, cl := range r.clients {
				if cl.IsBot() && r.IsGameOver() {
					continue
				}
				st, _ = r.gameBridge.Status(n)
				r.observer.Trigger(GameStatusUpdated, cl, st, r.updateSequenceNumber)
			}
			if r.turnMovedToNewPlayers() {
				r.changeClientsInTurn()
			}
		} else {
			response := messages.New(interfaces.TypeMessageError, err.Error())
			r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
		}
	}
}

func (r *Room) messageAuthorIsInTurn(m *interfaces.IncomingMessage) bool {
	for _, cl := range r.clientsInTurn {
		if m.Author == cl {
			return true
		}
	}
	return false
}

func (r *Room) turnMovedToNewPlayers() bool {
	gameCurrentPlayersClients, _ := r.gameCurrentPlayersClients()

	if len(gameCurrentPlayersClients) != len(r.clientsInTurn) {
		return true
	}

	for i := range gameCurrentPlayersClients {
		if gameCurrentPlayersClients[i] != r.clientsInTurn[i] {
			return true
		}
	}
	return false
}

func (r *Room) changeClientsInTurn() {
	for _, cl := range r.clientsInTurn {
		cl.StopTimer()
	}
	r.clientsInTurn, _ = r.gameCurrentPlayersClients()
	r.startClientsInTurnTimers()
}

func (r *Room) startClientsInTurnTimers() {
	for _, cl := range r.clientsInTurn {
		if !cl.IsBot() && r.playerTimeOut > 0 {
			cl.StartTimer(time.Second * r.playerTimeOut)
		}
	}
}

func (r *Room) playersData() map[string]interfaces.PlayerData {
	players := make(map[string]interfaces.PlayerData, len(r.clients))
	for n, c := range r.clients {
		players[strconv.Itoa(n)] = interfaces.PlayerData{
			Name: c.Name(),
		}
	}
	return players
}

func (r *Room) gameCurrentPlayersClients() ([]interfaces.Client, error) {
	currentPlayerClients := []interfaces.Client{}
	numbers, err := r.gameBridge.CurrentPlayersNumbers()
	for _, n := range numbers {
		currentPlayerClients = append(currentPlayerClients, r.clients[n])
	}
	return currentPlayerClients, err
}

// AddHuman adds a new client to the room. If the client has successfully joined,
// a message with his/her number in the room is send back to the client.
func (r *Room) AddHuman(cl interfaces.Client) error {
	var err error
	var clientNumber int

	if clientNumber, err = r.addClient(cl); err == nil {
		if r.configuration.Debug {
			log.Printf("Client '%s' added to room", cl.Name())
		}
		response := messages.New(interfaces.TypeMessageJoinedRoom, clientNumber, r.id, cl == r.owner)
		r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response, interfaces.TypeMessageJoinedRoom)
	}
	return err
}

func (r *Room) addClient(c interfaces.Client) (int, error) {
	mutex.Lock()
	defer mutex.Unlock()

	r.clients[r.clientCounter] = c
	newClientNumber := r.clientCounter
	r.clientCounter++
	if len(r.clients) == 1 {
		r.owner = c
	}
	c.SetRoom(r)
	response := messages.New(interfaces.TypeMessageCurrentPlayers, r.playersData())
	r.observer.Trigger("messageCreated", mapToSlice(r.clients), response, interfaces.TypeMessageCurrentPlayers)

	return newClientNumber, nil
}

// RemoveClient removes a client and its player
// depending wheter the game has already started or not.
func (r *Room) RemoveClient(c interfaces.Client) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := range r.clients {
		if r.clients[i] == c {
			r.clients[i].SetRoom(nil)
			c.StopTimer()

			delete(r.clients, i)

			if r.gameBridge.GameStarted() && !r.gameBridge.IsGameOver() {
				r.removePlayer(i)
			} else {
				response := messages.New(interfaces.TypeMessageCurrentPlayers, r.playersData())
				r.observer.Trigger("messageCreated", r.HumanClients(), response, interfaces.TypeMessageCurrentPlayers)
			}

			r.observer.Trigger(ClientOut, r)

			return
		}
	}
}

// removePlayer removes a player from a game,
// and returns an updated game status to all the players as a response
func (r *Room) removePlayer(playerNumber int) {
	r.gameBridge.RemovePlayer(playerNumber)

	if r.turnMovedToNewPlayers() {
		r.changeClientsInTurn()
	}

	r.updateSequenceNumber++
	for i, cl := range r.clients {
		if cl.IsBot() {
			continue
		}
		st, _ := r.gameBridge.Status(i)
		r.observer.Trigger(GameStatusUpdated, cl, st)
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
func (r *Room) Clients() map[int]interfaces.Client {
	mutex.Lock()
	defer mutex.Unlock()
	return r.clients
}

// HumanClients returns room's connected human clients
func (r *Room) HumanClients() []interfaces.Client {
	//mutex.Lock()
	//defer mutex.Unlock()
	human := []interfaces.Client{}
	for _, c := range r.clients {
		if !c.IsBot() {
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

func mapToSlice(in map[int]interfaces.Client) []interfaces.Client {
	var out []interfaces.Client
	for _, cl := range in {
		out = append(out, cl)
	}
	return out
}
