package interfaces

import (
	"encoding/json"
)

// These constants define the different websocket close codes that signal
// the reason why that websocket connection was closed. Note that, following
// WS standard, all codes goes from 4000 onwards
// https://developer.mozilla.org/es/docs/Web/API/CloseEvent
const (
	EndOk        = 4000
	HubDestroyed = 4001
	PlayerQuit   = 4002
	PlayerKicked = 4003
	HubTimeout   = 4004
)

// Message is a container struct used by
// clients to encapsulate their action messages sent to the hub.
type ClientMessage struct {
	Author  Client
	Content ClientMessageContent
}

// MessageContent is a struct that goes inside Message struct, acting
// as a container for the different kind of parameters needed for each type
// of message
type ClientMessageContent struct {
	Type   string          `json:"typ"`
	Room   string          `json:"rom"`
	Params json.RawMessage `json:"par"`
}

type Client interface {
	ReadPump(channel interface{}, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Name() string
	SetName(v string) Client
	Close(code int)
	IsBot() bool
	Room() Room
	SetRoom(r Room)
}
