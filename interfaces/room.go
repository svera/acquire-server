package interfaces

import "time"

// Room is an interface that defines the minimum set of functions a room
// implementation must have
type Room interface {
	GameStarted() bool
	ParseMessage(m *MessageFromClient) (map[Client][]byte, error)
	IsGameOver() bool
	RemoveClient(c Client) map[Client][]byte
	ID() string
	Owner() Client
	Clients() []Client
	HumanClients() []Client
	AddClient(c Client) (map[Client][]byte, error)
	SetTimer(t *time.Timer)
	Timer() *time.Timer
}
