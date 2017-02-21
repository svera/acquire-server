package interfaces

import "time"

// Room is an interface that defines the minimum set of functions a room
// implementation must have
type Room interface {
	GameStarted() bool
	Parse(m *IncomingMessage)
	IsGameOver() bool
	RemoveClient(c Client)
	ID() string
	Owner() Client
	Clients() map[int]Client
	HumanClients() []Client
	AddHuman(c Client) error
	SetTimer(t *time.Timer)
	Timer() *time.Timer
}
