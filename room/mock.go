package room

import (
	"time"

	"github.com/svera/sackson-server/interfaces"
)

// Mock is a structure that implements the Room interface for testing
type Mock struct {
	FakeGameStarted               func() bool
	FakeParse                     func(m *interfaces.IncomingMessage)
	FakeIsGameOver                func() bool
	FakeRemoveClient              func(c interfaces.Client)
	FakeID                        func() string
	FakeOwner                     func() interfaces.Client
	FakeClients                   func() map[int]interfaces.Client
	FakeHumanClients              func() []interfaces.Client
	FakeAddHuman                  func(c interfaces.Client) error
	FakeSetTimer                  func(t *time.Timer)
	FakeTimer                     func() *time.Timer
	FakeGameCurrentPlayersClients func() ([]interfaces.Client, error)
	FakeIsToBeDestroyed           func() bool
	FakeToBeDestroyed             func(bool)
	FakeGameDriverName            func() string
	FakePlayerTimeOut             func() time.Duration
	Calls                         map[string]int
}

// NewMock returns a new Mock instance ready to use
func NewMock() *Mock {
	return &Mock{
		FakeID: func() string {
			return "testRoom"
		},
		FakeSetTimer: func(t *time.Timer) {
			return
		},
		FakeGameDriverName: func() string {
			return "test"
		},
		FakeGameStarted: func() bool {
			return false
		},
		FakeAddHuman: func(c interfaces.Client) error {
			return nil
		},
		FakeToBeDestroyed: func(bool) {

		},
		FakeTimer: func() *time.Timer {
			return nil
		},
		FakeClients: func() map[int]interfaces.Client {
			return make(map[int]interfaces.Client)
		},
		FakeRemoveClient: func(c interfaces.Client) {
		},
		FakeHumanClients: func() []interfaces.Client {
			return make([]interfaces.Client, 0)
		},
		FakeIsToBeDestroyed: func() bool {
			return false
		},
		Calls: make(map[string]int),
	}
}

// GameStarted mocks the GameStarted method defined in the Room interface
func (r *Mock) GameStarted() bool {
	return r.FakeGameStarted()
}

// Parse mocks the Parse method defined in the Room interface
func (r *Mock) Parse(m *interfaces.IncomingMessage) {
	r.FakeParse(m)
}

// IsGameOver mocks the IsGameOver method defined in the Room interface
func (r *Mock) IsGameOver() bool {
	return r.FakeIsGameOver()
}

// RemoveClient mocks the RemoveClient method defined in the Room interface
func (r *Mock) RemoveClient(c interfaces.Client) {
	r.FakeRemoveClient(c)
}

// ID mocks the ID method defined in the Room interface
func (r *Mock) ID() string {
	return r.FakeID()
}

// Owner mocks the Owner method defined in the Room interface
func (r *Mock) Owner() interfaces.Client {
	return r.FakeOwner()
}

// Clients mocks the Clients method defined in the Room interface
func (r *Mock) Clients() map[int]interfaces.Client {
	return r.FakeClients()
}

// HumanClients mocks the HumanClients method defined in the Room interface
func (r *Mock) HumanClients() []interfaces.Client {
	return r.FakeHumanClients()
}

// AddHuman mocks the AddHuman method defined in the Room interface
func (r *Mock) AddHuman(c interfaces.Client) error {
	r.Calls["AddHuman"]++
	return r.FakeAddHuman(c)
}

// SetTimer mocks the SetTimer method defined in the Room interface
func (r *Mock) SetTimer(t *time.Timer) {
	r.FakeSetTimer(t)
}

// Timer mocks the Timer method defined in the Room interface
func (r *Mock) Timer() *time.Timer {
	return r.FakeTimer()
}

// GameCurrentPlayersClients mocks the GameCurrentPlayersClients method defined in the Room interface
func (r *Mock) GameCurrentPlayersClients() ([]interfaces.Client, error) {
	return r.FakeGameCurrentPlayersClients()
}

// IsToBeDestroyed mocks the IsToBeDestroyed method defined in the Room interface
func (r *Mock) IsToBeDestroyed() bool {
	return r.FakeIsToBeDestroyed()
}

// ToBeDestroyed mocks the ToBeDestroyed method defined in the Room interface
func (r *Mock) ToBeDestroyed(value bool) {
	r.FakeToBeDestroyed(value)
}

// GameDriverName mocks the GameDriverName method defined in the Room interface
func (r *Mock) GameDriverName() string {
	return r.FakeGameDriverName()
}

// PlayerTimeOut mocks the PlayerTimeOut method defined in the Room interface
func (r *Mock) PlayerTimeOut() time.Duration {
	return r.FakePlayerTimeOut()
}
