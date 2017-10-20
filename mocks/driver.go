package mocks

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

// Driver is a structure that implements the Driver interface for testing
type Driver struct {
	FakeCurrentPlayersNumbers []int
	FakeStatus                []byte
	FakeAI                    interfaces.AI
	FakeGameStarted           bool
	FakeIsGameOver            bool
	FakeExecute               func(clientName string, t string, content json.RawMessage) error
	Calls                     map[string]int
}

// Execute mocks the Execute method defined in the Driver interface
func (b *Driver) Execute(clientName string, t string, content json.RawMessage) error {
	if b.FakeExecute != nil {
		return b.FakeExecute(clientName, t, content)
	}
	return nil
}

// ParseMessage mocks the ParseMessage method defined in the Driver interface
func (b *Driver) ParseMessage(t string, content json.RawMessage) error {
	return nil
}

// CurrentPlayersNumbers mocks the CurrentPlayesrNumbers method defined in the Driver interface
func (b *Driver) CurrentPlayersNumbers() ([]int, error) {
	return b.FakeCurrentPlayersNumbers, nil
}

// Status mocks the Status method defined in the Driver interface
func (b *Driver) Status(playerNumber int) (interface{}, error) {
	return b.FakeStatus, nil
}

// RemovePlayer mocks the RemovePlayer method defined in the Driver interface
func (b *Driver) RemovePlayer(number int) error {
	return nil
}

// DeactivatePlayer mocks the DeactivatePlayer method defined in the Driver interface
func (b *Driver) DeactivatePlayer(number int) error {
	return nil
}

// CreateAI mocks the CreateAI method defined in the Driver interface
func (b *Driver) CreateAI(params interface{}) (interface{}, error) {
	return b.FakeAI, nil
}

// StartGame mocks the StartGame method defined in the Driver interface
func (b *Driver) StartGame(players map[int]string) error {
	b.Calls["StartGame"]++
	return nil
}

// GameStarted mocks the GameStarted method defined in the Driver interface
func (b *Driver) GameStarted() bool {
	return b.FakeGameStarted
}

// IsGameOver mocks the IsGameOver method defined in the Driver interface
func (b *Driver) IsGameOver() bool {
	return b.FakeIsGameOver
}

// Name mocks the Name method defined in the Driver interface
func (b *Driver) Name() string {
	return "mock"
}
