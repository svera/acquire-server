package interfaces

import "time"

type Room interface {
	GameStarted() bool
	ParseMessage(m *MessageFromClient) (map[Client][]byte, error)
	IsGameOver() bool
	RemoveClient(c Client) map[Client][]byte
	ID() string
	Owner() Client
	Clients() []Client
	AddClient(c Client) (map[Client][]byte, error)
	SetTimer(t *time.Timer)
	Timer() *time.Timer
}
