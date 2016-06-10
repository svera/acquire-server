package interfaces

import (
	"encoding/json"
)

// Bridge is an interface that defines the minimum set of functions needed
// to implement a game bridge which can be used within a hub instance
type Bridge interface {
	Execute(t string, content json.RawMessage) error
	CurrentPlayerNumber() (int, error)
	Status(n int) ([]byte, error)
	AddPlayer(name string) error
	RemovePlayer(number int) error
	DeactivatePlayer(number int) error
	AddBot(params interface{}) (Client, error)
	StartGame() error
	GameStarted() bool
	IsGameOver() bool
}
