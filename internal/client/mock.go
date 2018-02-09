package client

import (
	"time"

	"github.com/svera/sackson-server/internal/interfaces"
)

// Mock is a structure that implements the Client interface for testing
type Mock struct {
	FakeReadPump   func(channel interface{}, unregister chan interfaces.Client)
	FakeWritePump  func()
	FakeIncoming   func() chan []byte
	FakeOwner      func() bool
	FakeSetOwner   func(bool) interfaces.Client
	FakeClose      func()
	FakeName       func() string
	FakeSetName    func(string) interfaces.Client
	FakeIsBot      func() bool
	FakeRoom       func() interfaces.Room
	FakeSetRoom    func(interfaces.Room)
	FakeSetTimer   func(t *time.Timer)
	FakeStopTimer  func()
	FakeStartTimer func(time.Duration)
	FakeSetGame    func(string)
	FakeGame       func() string
}

// NewMock returns a new mock instance ready to use
func NewMock() *Mock {
	c := &Mock{
		FakeIncoming: func() chan []byte {
			return make(chan []byte, 2)
		},
		FakeName: func() string {
			return "TestClient"
		},
		FakeGame: func() string {
			return "test"
		},
		FakeClose: func() {
			// Do nothing
		},
		FakeStopTimer: func() {
			// Do nothing
		},
		FakeRoom: func() interfaces.Room {
			return nil
		},
		FakeSetRoom: func(interfaces.Room) {

		},
	}

	c.FakeWritePump = func() {
		for range c.Incoming() {
			// Do nothing
		}
	}

	c.FakeSetName = func(string) interfaces.Client {
		return c
	}
	return c
}

// ReadPump mocks the ReadPump method defined in the Client interface
func (c *Mock) ReadPump(channel interface{}, unregister chan interfaces.Client) {
	c.FakeReadPump(channel, unregister)
}

// WritePump mocks the WritePump method defined in the Client interface
func (c *Mock) WritePump() {
	c.FakeWritePump()
}

// Incoming mocks the Incoming method defined in the Client interface
func (c *Mock) Incoming() chan []byte {
	return c.FakeIncoming()
}

// Owner mocks the Owner method defined in the Client interface
func (c *Mock) Owner() bool {
	return c.FakeOwner()
}

// SetOwner mocks the SetOwner method defined in the Client interface
func (c *Mock) SetOwner(v bool) interfaces.Client {
	return c.FakeSetOwner(v)
}

// Name mocks the Name method defined in the Client interface
func (c *Mock) Name() string {
	return c.FakeName()
}

// SetName mocks the SetName method defined in the Client interface
func (c *Mock) SetName(v string) interfaces.Client {
	return c.FakeSetName(v)
}

// Close mocks the Close method defined in the Client interface
func (c *Mock) Close() {
	c.FakeClose()
}

// Room mocks the Room method defined in the Client interface
func (c *Mock) Room() interfaces.Room {
	return c.FakeRoom()
}

// SetRoom mocks the SetRoom method defined in the Client interface
func (c *Mock) SetRoom(r interfaces.Room) {
	c.FakeSetRoom(r)
}

// IsBot mocks the IsBot method defined in the Client interface
func (c *Mock) IsBot() bool {
	return c.FakeIsBot()
}

// SetTimer mocks the IsBot method defined in the Client interface
func (c *Mock) SetTimer(t *time.Timer) {
	c.FakeSetTimer(t)
}

// StopTimer mocks the IsBot method defined in the Client interface
func (c *Mock) StopTimer() {
	c.FakeStopTimer()
}

// StartTimer mocks the IsBot method defined in the Client interface
func (c *Mock) StartTimer(d time.Duration) {
	c.FakeStartTimer(d)
}

// SetGame mocks the SetGame method defined in the Client interface
func (c *Mock) SetGame(game string) {
	c.FakeSetGame(game)
}

// Game mocks the Game method defined in the Client interface
func (c *Mock) Game() string {
	return c.FakeGame()
}
