package hub

import (
	"encoding/json"
)

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
	GameFull           = "game_full"
	GameNotStarted     = "game_not_started"
	GameAlreadyStarted = "game_already_started"
	WrongMessage       = "message_parsing_error"
	NotEnoughPlayers   = "not_enough_players"
)
