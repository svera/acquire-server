// Package room contains the Room class, which manages communication between clients and game,
// passing messages back and forth which describe actions and results,
// as well as the connections to it.
package room

import (
	"encoding/json"
	"errors"
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
	// Registered clients
	clients []interfaces.Client

	owner interfaces.Client

	// Stops room server
	stop chan struct{}

	gameBridge interfaces.Bridge

	// Maximum time this room instance will be kept alive
	timeout time.Duration

	wasClosedByTimeout bool
}

// New returns a new Room instance
func New(b interfaces.Bridge, cfg *config.Config) *Room {
	return &Room{
		stop:               make(chan struct{}),
		clients:            []interfaces.Client{},
		gameBridge:         b,
		timeout:            cfg.Timeout,
		wasClosedByTimeout: false,
	}
}

func (r *Room) AddBot(level string) (interfaces.Client, error) {
	var err error
	var c interfaces.Client
	if c, err = r.gameBridge.AddBot(level); err == nil {
		if err = r.AddClient(c); err == nil {
			return c, nil
		}
	}
	return nil, err
}

func (r *Room) ParseMessage(m *interfaces.ClientMessage) error {
	var err error
	var currentPlayer interfaces.Client

	if r.gameBridge.IsGameOver() {
		return errors.New(GameOver)
	}

	if r.isControlMessage(m) {
		err = r.parseControlMessage(m)
	} else {
		if currentPlayer, err = r.currentPlayerClient(); m.Author == currentPlayer && err == nil {
			err = r.gameBridge.Execute(m.Content.Type, m.Content.Params)
		}
	}
	if err != nil {
		log.Println(err)
	}
	return err
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

func (r *Room) parseControlMessage(m *interfaces.ClientMessage) error {
	if m.Author != r.Owner() {
		return nil
	}
	switch m.Content.Type {

	case client.ControlMessageTypeStartGame:
		if err := r.StartGame(); err != nil {
			return err
		}

	case client.ControlMessageTypeJoinPlayer:
		if err := r.AddClient(m.Author); err != nil {
			return err
		}

	case client.ControlMessageTypeAddBot:
		var parsed client.AddBotMessageParams
		if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
			if cl, err := r.AddBot(parsed.BotName); err != nil {
				return err
			} else {
				_ = cl
				//go cl.WritePump()
				//go cl.ReadPump(h.Messages, h.Unregister)
			}
		}
		/*
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
		*/
	}
	return nil
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

func (r *Room) StartGame() error {
	return r.gameBridge.StartGame()
}

func (r *Room) AddClient(c interfaces.Client) error {
	if err := r.gameBridge.AddPlayer(c.Name()); err != nil {
		return err
	}
	r.clients = append(r.clients, c)

	if len(r.clients) == 1 {
		r.owner = c
	}
	return nil
}

func (r *Room) Owner() interfaces.Client {
	return r.owner
}

func (r *Room) kickClient(number int) error {
	if number < 0 || number > len(r.clients) {
		return errors.New(InexistentClient)
	}
	if r.clients[number] == r.Owner() {
		return errors.New(OwnerNotRemovable)
	}
	r.clients[number].Close(interfaces.PlayerKicked)
	r.removeClient(r.clients[number])
	return nil
}

func (r *Room) quitClient(client interfaces.Client) error {
	if client == r.Owner() {
		return errors.New(OwnerNotRemovable)
	}
	client.Close(interfaces.PlayerQuit)
	r.removeClient(client)
	return nil
}

func (r *Room) terminateGame(client interfaces.Client) error {
	if client != r.Owner() {
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

	close(r.stop)
}

// Removes /sets as nil a client and removes / deactivates its player
// depending wheter the game has already started or not.
// Note that we don't remove a client if a game has already started, as client
// indexes must not change once a game has started.
func (r *Room) removeClient(c interfaces.Client) {
	for i := range r.clients {
		if r.clients[i] == c {
			if r.gameBridge.GameStarted() {
				r.clients[i] = nil
				r.gameBridge.DeactivatePlayer(i)
				//r.broadcastUpdate()
			} else {
				r.clients = append(r.clients[:i], r.clients[i+1:]...)
				r.gameBridge.RemovePlayer(i)
				//r.sendUpdatedPlayersList()
			}
			log.Printf("Cliente eliminado, Numero de clientes: %d\n", len(r.clients))
			return
		}
	}
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

func (r *Room) Status(n int) ([]byte, error) {
	return r.gameBridge.Status(n)
}
