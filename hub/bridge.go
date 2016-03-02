package hub

import (
	"encoding/json"
)

// Bridge is an interface that defines the minimum set of functions needed
// to implement a game bridge which can be used within a hub instance
type Bridge interface {
	ParseMessage(t string, content json.RawMessage) ([]byte, error)
	CurrentPlayerNumber() (int, error)
	Status(n int) ([]byte, error)
	AddPlayer() error
	StartGame() error
}
