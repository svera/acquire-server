package drivers

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

// Mock is a structure that implements the Driver interface for testing
type Mock struct {
	FakeCurrentPlayersNumbers []int
	FakeStatus                []byte
	FakeAI                    interfaces.AI
	FakeGameStarted           bool
	FakeIsGameOver            bool
	FakeExecute               func(clientName string, t string, content json.RawMessage) error
	Calls                     map[string]int
}

// NewMock returns a new Mock instance ready to use
func NewMock() *Mock {
	return &Mock{
		FakeAI: &AI{},
		Calls:  make(map[string]int),
	}
}

// Execute mocks the Execute method defined in the Driver interface
func (b *Mock) Execute(clientName string, t string, content json.RawMessage) error {
	if b.FakeExecute != nil {
		return b.FakeExecute(clientName, t, content)
	}
	return nil
}

// ParseMessage mocks the ParseMessage method defined in the Driver interface
func (b *Mock) ParseMessage(t string, content json.RawMessage) error {
	return nil
}

// CurrentPlayersNumbers mocks the CurrentPlayesrNumbers method defined in the Driver interface
func (b *Mock) CurrentPlayersNumbers() ([]int, error) {
	return b.FakeCurrentPlayersNumbers, nil
}

// Status mocks the Status method defined in the Driver interface
func (b *Mock) Status(playerNumber int) (interface{}, error) {
	return b.FakeStatus, nil
}

// RemovePlayer mocks the RemovePlayer method defined in the Driver interface
func (b *Mock) RemovePlayer(number int) error {
	return nil
}

// DeactivatePlayer mocks the DeactivatePlayer method defined in the Driver interface
func (b *Mock) DeactivatePlayer(number int) error {
	return nil
}

// CreateAI mocks the CreateAI method defined in the Driver interface
func (b *Mock) CreateAI(params interface{}) (interface{}, error) {
	return b.FakeAI, nil
}

// StartGame mocks the StartGame method defined in the Driver interface
func (b *Mock) StartGame(players map[int]string) error {
	b.Calls["StartGame"]++
	return nil
}

// GameStarted mocks the GameStarted method defined in the Driver interface
func (b *Mock) GameStarted() bool {
	return b.FakeGameStarted
}

// IsGameOver mocks the IsGameOver method defined in the Driver interface
func (b *Mock) IsGameOver() bool {
	return b.FakeIsGameOver
}

// Name mocks the Name method defined in the Driver interface
func (b *Mock) Name() string {
	return "mock"
}
