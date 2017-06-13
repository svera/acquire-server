package interfaces

import (
	"encoding/json"
)

// Bridge is an interface that defines the minimum set of functions needed
// to implement a game bridge which can be used within a hub instance
type Bridge interface {
	Execute(clientName string, t string, content json.RawMessage) error
	CurrentPlayersNumbers() ([]int, error)
	Status(n int) (interface{}, error)
	RemovePlayer(number int) error
	AddBot(params interface{}, room Room) (Client, error)
	StartGame(map[int]Client) error
	GameStarted() bool
	IsGameOver() bool
}
