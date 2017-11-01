package room

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

var (
	mutex sync.RWMutex
)

// Room is a struct that manage the message flow between client (players)
// and a game. It can work with any game as long as it implements the Driver
// interface. It also provides support for some common operations as adding/removing
// players and more.
type Room struct {
	id string

	// Registered clients
	clients map[int]interfaces.Client

	owner interfaces.Client

	gameDriver interfaces.Driver

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

	toBeDestroyed bool
}

// New returns a new Room instance
func New(
	id string,
	g interfaces.Driver,
	owner interfaces.Client,
	messages chan *interfaces.IncomingMessage,
	unregister chan interfaces.Client,
	cfg *config.Config,
	ob interfaces.Observer,
) *Room {
	return &Room{
		id:                   id,
		clients:              map[int]interfaces.Client{},
		gameDriver:           g,
		owner:                owner,
		messages:             messages,
		unregister:           unregister,
		observer:             ob,
		clientsInTurn:        nil,
		configuration:        cfg,
		clientCounter:        0,
		updateSequenceNumber: 0,
		toBeDestroyed:        false,
	}
}

// Parse gets an incoming message from a client and parses it, executing
// its desired action in the room or passing it to the room's game driver
func (r *Room) Parse(m *interfaces.IncomingMessage) {
	if r.isControlMessage(m) {
		r.parseControlMessage(m)
	} else if r.gameDriver.IsGameOver() {
		r.observer.Trigger(events.Error{Client: m.Author, ErrorText: GameOver})
	} else {
		r.passMessageToGame(m)
	}
}

func (r *Room) isControlMessage(m *interfaces.IncomingMessage) bool {
	switch m.Type {
	case
		messages.TypeAddBot,
		messages.TypeStartGame,
		messages.TypeKickPlayer,
		messages.TypePlayerQuits,
		messages.TypeSetClientData:
		return true
	}
	return false
}

func (r *Room) parseControlMessage(m *interfaces.IncomingMessage) {
	var err error
	switch m.Type {

	case messages.TypeStartGame:
		err = r.startGameAction(m)

	case messages.TypeAddBot:
		err = r.addBotAction(m)

	case messages.TypeKickPlayer:
		err = r.kickPlayerAction(m)

	case messages.TypePlayerQuits:
		err = r.clientQuits(m.Author)

	case messages.TypeSetClientData:
		err = r.setClientDataAction(m)
	}

	if err != nil {
		r.observer.Trigger(events.Error{Client: m.Author, ErrorText: err.Error()})
	}
}

func (r *Room) passMessageToGame(m *interfaces.IncomingMessage) {
	var err error
	var st interface{}

	if r.messageAuthorIsInTurn(m) {
		if err = r.gameDriver.Execute(m.Author.Name(), m.Type, m.Content); err == nil {
			r.updateSequenceNumber++
			for n, cl := range r.clients {
				if cl.IsBot() && r.IsGameOver() {
					continue
				}
				st, _ = r.gameDriver.Status(n)
				r.observer.Trigger(events.GameStatusUpdated{Client: cl, Message: st, SequenceNumber: r.updateSequenceNumber})
			}
			if r.turnMovedToNewPlayers() {
				r.changeClientsInTurn()
			}
		} else {
			r.observer.Trigger(events.Error{Client: m.Author, ErrorText: err.Error()})
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
	gameCurrentPlayersClients, _ := r.GameCurrentPlayersClients()

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
	r.clientsInTurn, _ = r.GameCurrentPlayersClients()
	r.startClientsInTurnTimers()
}

func (r *Room) startClientsInTurnTimers() {
	for _, cl := range r.clientsInTurn {
		if !cl.IsBot() && r.playerTimeOut > 0 {
			cl.StartTimer(time.Second * r.playerTimeOut)
		}
	}
}

func (r *Room) playersData() map[string]messages.PlayerData {
	players := make(map[string]messages.PlayerData, len(r.clients))
	for n, c := range r.clients {
		players[strconv.Itoa(n)] = messages.PlayerData{
			Name: c.Name(),
		}
	}
	return players
}

func (r *Room) GameCurrentPlayersClients() ([]interfaces.Client, error) {
	currentPlayerClients := []interfaces.Client{}
	numbers, err := r.gameDriver.CurrentPlayersNumbers()
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
		r.observer.Trigger(events.ClientJoined{Client: cl, ClientNumber: clientNumber, Owner: cl == r.owner})
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
	r.observer.Trigger(events.ClientsUpdated{Clients: mapToSlice(r.clients), PlayersData: r.playersData()})

	return newClientNumber, nil
}

// RemoveClient removes a client and its player
// depending wether the game has already started or not.
func (r *Room) RemoveClient(c interfaces.Client) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := range r.clients {
		if r.clients[i] == c {
			r.clients[i].SetRoom(nil)
			c.StopTimer()
			delete(r.clients, i)

			if len(r.HumanClients()) == 0 {
				return
			}

			if r.gameDriver.GameStarted() && !r.gameDriver.IsGameOver() {
				r.removePlayer(i)
			} else {
				r.observer.Trigger(events.ClientsUpdated{Clients: r.HumanClients(), PlayersData: r.playersData()})
			}

			return
		}
	}
}

// removePlayer removes a player from a game,
// and returns an updated game status to all the players as a response
func (r *Room) removePlayer(playerNumber int) {
	r.gameDriver.RemovePlayer(playerNumber)

	if r.turnMovedToNewPlayers() {
		r.changeClientsInTurn()
	}

	r.updateSequenceNumber++
	for i, cl := range r.clients {
		st, _ := r.gameDriver.Status(i)
		r.observer.Trigger(events.GameStatusUpdated{Client: cl, Message: st, SequenceNumber: r.updateSequenceNumber})
	}
}

// GameStarted returns true if the room's game has started, false otherwise
func (r *Room) GameStarted() bool {
	return r.gameDriver.GameStarted()
}

// IsGameOver returns true if the room's game has ended, false otherwise
func (r *Room) IsGameOver() bool {
	return r.gameDriver.IsGameOver()
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

// IsToBeDestroyed returns true if the room has been marked to be destroyed, false otherwise
func (r *Room) IsToBeDestroyed() bool {
	return r.toBeDestroyed
}

// ToBeDestroyed sets wether a room has to be destroyed or not
func (r *Room) ToBeDestroyed(value bool) {
	r.toBeDestroyed = value
}

// GameDriverName returns the name of the game driver being used by the room
func (r *Room) GameDriverName() string {
	return r.gameDriver.Name()
}

// PlayerTimeOut returns the allowed time per turn for every player
func (r *Room) PlayerTimeOut() time.Duration {
	return r.playerTimeOut
}
