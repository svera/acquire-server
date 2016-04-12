package client

import (
	"encoding/json"

	"github.com/svera/tbg-server/interfaces"
)

// Control messages types, common to all games.
// Game-specific messages are defined at game bridge level.
const (
	ControlMessageTypeStartGame = "ini"
	ControlMessageTypeAddBot    = "bot"
)

type MessageContent struct {
	Type   string          `json:"typ"`
	Params json.RawMessage `json:"par"`
}

// Message is a container struct used by
// clients to encapsulate their action messages sent to the hub.
type Message struct {
	Author  interfaces.Client
	Content MessageContent
}
