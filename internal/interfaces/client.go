package interfaces

import "time"

// Client is an interface that defines the minimum set of functions needed
// to implement a client which can be used within a hub instance
type Client interface {
	ReadPump(channel chan *IncomingMessage, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Name() string
	SetName(v string) Client
	Close()
	IsBot() bool
	Room() Room
	SetRoom(r Room)
	SetTimer(t *time.Timer)
	StopTimer()
	StartTimer(d time.Duration)
	SetGame(game string)
	Game() string
}
