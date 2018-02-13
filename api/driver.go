package api

import (
	"encoding/json"
)

// Driver is an interface that defines the minimum set of functions needed
// to implement a game driver which can be used within a hub instance
type Driver interface {
	Execute(clientName string, messageType string, content json.RawMessage) error
	CurrentPlayersNumbers() ([]int, error)
	Status(playerNumber int) (interface{}, error)
	RemovePlayer(number int) error
	CreateAI(params interface{}) (AI, error)
	StartGame(clientNames map[int]string) error
	GameStarted() bool
	IsGameOver() bool
	Name() string
}
