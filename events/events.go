package events

import "github.com/svera/sackson-server/interfaces"

type GameStarted struct {
}

type ClientRegistered struct {
	Client interfaces.Client
}

type ClientUnregistered struct {
	Client interfaces.Client
}

type ClientOut struct {
	Client interfaces.Client
	Reason string
	Room   interfaces.Room
}

type ClientJoined struct {
	Client       interfaces.Client
	ClientNumber int
	Owner        bool
}

type ClientsUpdated struct {
	Clients     []interfaces.Client
	PlayersData map[string]interfaces.PlayerData
}

type GameStatusUpdated struct {
	Client         interfaces.Client
	Message        interface{}
	SequenceNumber int
}

type RoomCreated struct {
}

type RoomDestroyed struct {
}

type Error struct {
	Client    interfaces.Client
	ErrorText string
}
