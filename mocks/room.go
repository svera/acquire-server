package mocks

import (
	"time"

	"github.com/svera/sackson-server/interfaces"
)

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

func (r *Room) GameStarted() bool {
	return r.FakeGameStarted()
}

func (r *Room) Parse(m *interfaces.IncomingMessage) {
	r.FakeParse(m)
}

func (r *Room) IsGameOver() bool {
	return r.FakeIsGameOver()
}

func (r *Room) RemoveClient(c interfaces.Client) {
	r.FakeRemoveClient(c)
}

func (r *Room) ID() string {
	return r.FakeID()
}

func (r *Room) Owner() interfaces.Client {
	return r.FakeOwner()
}

func (r *Room) Clients() map[int]interfaces.Client {
	return r.FakeClients()
}

func (r *Room) HumanClients() []interfaces.Client {
	return r.FakeHumanClients()
}

func (r *Room) AddHuman(c interfaces.Client) error {
	return r.FakeAddHuman(c)
}

func (r *Room) SetTimer(t *time.Timer) {
	r.FakeSetTimer(t)
}

func (r *Room) Timer() *time.Timer {
	return r.FakeTimer()
}

func (r *Room) GameCurrentPlayersClients() ([]interfaces.Client, error) {
	return r.FakeGameCurrentPlayersClients()
}

func (r *Room) IsToBeDestroyed() bool {
	return r.FakeIsToBeDestroyed()
}

func (r *Room) ToBeDestroyed(value bool) {
	r.ToBeDestroyed(value)
}

func (r *Room) GameDriverName() string {
	return r.FakeGameDriverName()
}

func (r *Room) PlayerTimeOut() time.Duration {
	return r.FakePlayerTimeOut()
}
