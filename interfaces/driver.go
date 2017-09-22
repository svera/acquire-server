package interfaces

import (
	"encoding/json"
)

// Driver is an interface that defines the minimum set of functions needed
// to implement a game driver which can be used within a hub instance
type Driver interface {
	Execute(clientName string, t string, content json.RawMessage) error
	CurrentPlayersNumbers() ([]int, error)
	Status(playerNumber int) (interface{}, error)
	RemovePlayer(number int) error
	// The returned interface{} value must implement the AI interface,
	// as defined in interfaces/ai.go
	CreateAI(params interface{}) (interface{}, error)
	StartGame(map[int]string) error
	GameStarted() bool
	IsGameOver() bool
}
