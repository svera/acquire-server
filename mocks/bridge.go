package mocks

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

// Bridge is a structure that implements the Bridge interface for testing
type Bridge struct {
	FakeCurrentPlayerNumber int
	FakeStatus              []byte
	FakeClient              interfaces.Client
	FakeGameStarted         bool
	FakeIsGameOver          bool
}

// ParseMessage mocks the ParseMessage method defined in the Bridge interface
func (b *Bridge) ParseMessage(t string, content json.RawMessage) error {
	return nil
}

// CurrentPlayerNumber mocks the CurrentPlayerNumber method defined in the Bridge interface
func (b *Bridge) CurrentPlayerNumber() (int, error) {
	return b.FakeCurrentPlayerNumber, nil
}

// Status mocks the Status method defined in the Bridge interface
func (b *Bridge) Status(n int) ([]byte, error) {
	return b.FakeStatus, nil
}

// AddPlayer mocks the AddPlayer method defined in the Bridge interface
func (b *Bridge) AddPlayer(name string) error {
	return nil
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
func (b *Bridge) AddBot(params interface{}) (interfaces.Client, error) {
	return b.FakeClient, nil
}

// StartGame mocks the StartGame method defined in the Bridge interface
func (b *Bridge) StartGame() error {
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
