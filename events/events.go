package events

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

// GameStarted is an event triggered when a game starts
type GameStarted struct {
	Room           interfaces.Room
	GameParameters json.RawMessage
}

// ClientRegistered is an event triggered when a client connects
type ClientRegistered struct {
	Client interfaces.Client
}

// ClientUnregistered is an event triggered when a client disconnects
type ClientUnregistered struct {
	Client interfaces.Client
}

// ClientOut is an event triggered when a client lefts a room
type ClientOut struct {
	Client interfaces.Client
	Reason string
	Room   interfaces.Room
}

// ClientJoined is an event triggered when a client joins a room
type ClientJoined struct {
	Client       interfaces.Client
	ClientNumber int
	Owner        bool
}

// ClientsUpdated is an event triggered when a client joins/lefts a room
type ClientsUpdated struct {
	Clients     []interfaces.Client
	PlayersData map[string]messages.PlayerData
}

// GameStatusUpdated is an event triggered when a game driver sends updates its state
type GameStatusUpdated struct {
	Client         interfaces.Client
	Message        interface{}
	SequenceNumber int
}

// RoomCreated is an event triggered when a room is created
type RoomCreated struct {
	Room interfaces.Room
}

// RoomDestroyed is an event triggered when a room is destroyed
type RoomDestroyed struct {
	GameName string
}

type BotPanicked struct {
	Client interfaces.Client
}

// Error is an event triggered when an error happens
type Error struct {
	Client    interfaces.Client
	ErrorText string
}
