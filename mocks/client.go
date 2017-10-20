package mocks

import (
	"sync"
	"time"

	"github.com/svera/sackson-server/interfaces"
)

var (
	mutex sync.RWMutex
)

// Client is a structure that implements the Client interface for testing
type Client struct {
	FakeOwner    bool
	FakeName     string
	FakeIsBot    bool
	FakeIncoming chan []byte
	FakeRoom     interfaces.Room
	FakeGame     string
}

// ReadPump mocks the ReadPump method defined in the Client interface
func (c *Client) ReadPump(channel interface{}, unregister chan interfaces.Client) {

}

// WritePump mocks the WritePump method defined in the Client interface
func (c *Client) WritePump() {
	for range c.FakeIncoming {
		// Do nothing
	}
}

// Incoming mocks the Incoming method defined in the Client interface
func (c *Client) Incoming() chan []byte {
	return c.FakeIncoming
}

// Owner mocks the Owner method defined in the Client interface
func (c *Client) Owner() bool {
	return c.FakeOwner
}

// SetOwner mocks the SetOwner method defined in the Client interface
func (c *Client) SetOwner(v bool) interfaces.Client {
	return c
}

// Name mocks the Name method defined in the Client interface
func (c *Client) Name() string {
	return c.FakeName
}

// SetName mocks the SetName method defined in the Client interface
func (c *Client) SetName(v string) interfaces.Client {
	return c
}

// Close mocks the Close method defined in the Client interface
func (c *Client) Close() {
	close(c.FakeIncoming)
}

// Room mocks the Room method defined in the Client interface
func (c *Client) Room() interfaces.Room {
	return c.FakeRoom
}

// SetRoom mocks the SetRoom method defined in the Client interface
func (c *Client) SetRoom(r interfaces.Room) {
	c.FakeRoom = r
}

// IsBot mocks the IsBot method defined in the Client interface
func (c *Client) IsBot() bool {
	return c.FakeIsBot
}

// SetTimer mocks the IsBot method defined in the Client interface
func (c *Client) SetTimer(t *time.Timer) {

}

// StopTimer mocks the IsBot method defined in the Client interface
func (c *Client) StopTimer() {

}

// StartTimer mocks the IsBot method defined in the Client interface
func (c *Client) StartTimer(d time.Duration) {

}

// SetGame mocks the SetGame method defined in the Client interface
func (c *Client) SetGame(game string) {
}

// Game mocks the Game method defined in the Client interface
func (c *Client) Game() string {
	return c.FakeGame
}
