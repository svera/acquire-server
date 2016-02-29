package hub

import (
	"encoding/json"
)

// Bridge is an interface that defines the minimum set of functions needed
// to implement a game bridge which can be used within a hub instance
type Bridge interface {
	ParseMessage(t string, content json.RawMessage) ([]byte, error)
	CurrentPlayerNumber() int
	Status(n int) []byte
	AddPlayer()
	GameStarted() bool
	StartGame() error
	MaximumPlayers() int
	MinimumPlayers() int
}

const (
	// GameFull is an error returned when a game already has the maximum number of players
	GameFull = "game_full"
	// GameNotStarted is an error returned when a player tries to do an action that requires a running game
	GameNotStarted = "game_not_started"
	// GameAlreadyStarted is an error returned when a player tries to start a game in a hub instance which an already running one
	GameAlreadyStarted = "game_already_started"
	// WrongMessage is an error returned when the hub cannot parse an incoming message
	WrongMessage = "message_parsing_error"
	// NotEnoughPlayers is an error message returned when the game manager tries to start a new game with not enough players
	NotEnoughPlayers = "not_enough_players"
)
