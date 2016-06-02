package mocks

import (
	"fmt"
	"time"

	"github.com/svera/tbg-server/interfaces"
)

// Client is a structure that implements the Client interface for testing
type Client struct {
	FakeOwner    bool
	FakeName     string
	FakeIsBot    bool
	FakeIncoming chan []byte
}

// ReadPump mocks the ReadPump method defined in the Client interface
func (c *Client) ReadPump(channel interface{}, unregister chan interfaces.Client) {

}

// WritePump mocks the WritePump method defined in the Client interface
func (c *Client) WritePump() {
	time.AfterFunc(time.Second*5, func() {
		return
	})

	for {
		select {

		case msg := <-c.FakeIncoming:
			fmt.Println(msg)
		}
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
func (c *Client) Close(code int) {

}

// IsBot mocks the IsBot method defined in the Client interface
func (c *Client) IsBot() bool {
	return c.FakeIsBot
}
