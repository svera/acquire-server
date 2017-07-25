package mocks

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

// Bridge is a structure that implements the Bridge interface for testing
type Bridge struct {
	FakeCurrentPlayersNumbers []int
	FakeStatus                []byte
	FakeClient                interfaces.Client
	FakeGameStarted           bool
	FakeIsGameOver            bool
	FakeExecute               func(clientName string, t string, content json.RawMessage) error
	Calls                     map[string]int
}

// Execute mocks the Execute method defined in the Bridge interface
func (b *Bridge) Execute(clientName string, t string, content json.RawMessage) error {
	if b.FakeExecute != nil {
		return b.FakeExecute(clientName, t, content)
	}
	return nil
}

// ParseMessage mocks the ParseMessage method defined in the Bridge interface
func (b *Bridge) ParseMessage(t string, content json.RawMessage) error {
	return nil
}

// CurrentPlayersNumbers mocks the CurrentPlayesrNumbers method defined in the Bridge interface
func (b *Bridge) CurrentPlayersNumbers() ([]int, error) {
	return b.FakeCurrentPlayersNumbers, nil
}

// Status mocks the Status method defined in the Bridge interface
func (b *Bridge) Status(playerNumber int) (interface{}, error) {
	return b.FakeStatus, nil
}

// RemovePlayer mocks the RemovePlayer method defined in the Bridge interface
func (b *Bridge) RemovePlayer(number int) error {
	return nil
}

// DeactivatePlayer mocks the DeactivatePlayer method defined in the Bridge interface
func (b *Bridge) DeactivatePlayer(number int) error {
	return nil
}

// AddBot mocks the AddBot method defined in the Bridge interface
func (b *Bridge) AddBot(params interface{}, room interfaces.Room) (interfaces.Client, error) {
	return b.FakeClient, nil
}

// StartGame mocks the StartGame method defined in the Bridge interface
func (b *Bridge) StartGame(players map[int]interfaces.Client) error {
	b.Calls["StartGame"]++
	return nil
}

// GameStarted mocks the GameStarted method defined in the Bridge interface
func (b *Bridge) GameStarted() bool {
	return b.FakeGameStarted
}

// IsGameOver mocks the IsGameOver method defined in the Bridge interface
func (b *Bridge) IsGameOver() bool {
	return b.FakeIsGameOver
}
