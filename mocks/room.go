package mocks

import (
	"time"

	"github.com/svera/sackson-server/interfaces"
)

// Room is a structure that implements the Room interface for testing
type Room struct {
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
}

// GameStarted mocks the GameStarted method defined in the Room interface
func (r *Room) GameStarted() bool {
	return r.FakeGameStarted()
}

// Parse mocks the Parse method defined in the Room interface
func (r *Room) Parse(m *interfaces.IncomingMessage) {
	r.FakeParse(m)
}

// IsGameOver mocks the IsGameOver method defined in the Room interface
func (r *Room) IsGameOver() bool {
	return r.FakeIsGameOver()
}

// RemoveClient mocks the RemoveClient method defined in the Room interface
func (r *Room) RemoveClient(c interfaces.Client) {
	r.FakeRemoveClient(c)
}

// ID mocks the ID method defined in the Room interface
func (r *Room) ID() string {
	return r.FakeID()
}

// Owner mocks the Owner method defined in the Room interface
func (r *Room) Owner() interfaces.Client {
	return r.FakeOwner()
}

// Clients mocks the Clients method defined in the Room interface
func (r *Room) Clients() map[int]interfaces.Client {
	return r.FakeClients()
}

// HumanClients mocks the HumanClients method defined in the Room interface
func (r *Room) HumanClients() []interfaces.Client {
	return r.FakeHumanClients()
}

// AddHuman mocks the AddHuman method defined in the Room interface
func (r *Room) AddHuman(c interfaces.Client) error {
	return r.FakeAddHuman(c)
}

// SetTimer mocks the SetTimer method defined in the Room interface
func (r *Room) SetTimer(t *time.Timer) {
	r.FakeSetTimer(t)
}

// Timer mocks the Timer method defined in the Room interface
func (r *Room) Timer() *time.Timer {
	return r.FakeTimer()
}

// GameCurrentPlayersClients mocks the GameCurrentPlayersClients method defined in the Room interface
func (r *Room) GameCurrentPlayersClients() ([]interfaces.Client, error) {
	return r.FakeGameCurrentPlayersClients()
}

// IsToBeDestroyed mocks the IsToBeDestroyed method defined in the Room interface
func (r *Room) IsToBeDestroyed() bool {
	return r.FakeIsToBeDestroyed()
}

// ToBeDestroyed mocks the ToBeDestroyed method defined in the Room interface
func (r *Room) ToBeDestroyed(value bool) {
	r.ToBeDestroyed(value)
}

// GameDriverName mocks the GameDriverName method defined in the Room interface
func (r *Room) GameDriverName() string {
	return r.FakeGameDriverName()
}

// PlayerTimeOut mocks the PlayerTimeOut method defined in the Room interface
func (r *Room) PlayerTimeOut() time.Duration {
	return r.FakePlayerTimeOut()
}
