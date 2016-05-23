package client

import (
	"encoding/json"

	"github.com/svera/tbg-server/interfaces"
)

// Control messages types, common to all games.
// Game-specific messages are defined at game bridge level.
const (
	ControlMessageTypeStartGame     = "ini"
	ControlMessageTypeAddBot        = "bot"
	ControlMessageTypeKickPlayer    = "kck"
	ControlMessageTypePlayerQuits   = "qui"
	ControlMessageTypeTerminateGame = "ter"
)

// MessageContent is a struct that goes inside Message struct, acting
// as a container for the different kind of parameters needed for each type
// of message
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

// KickPlayerMessageParams contains the needed parameters for a kick player
// message
type KickPlayerMessageParams struct {
	PlayerNumber int `json:"ply"`
}

// AddBotMessageParams contains the needed parameters for a add bot
// message
type AddBotMessageParams struct {
	BotName string `json:"nam"`
}
