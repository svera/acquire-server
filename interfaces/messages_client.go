package interfaces

import (
	"encoding/json"
)

// Control messages types, common to all games.
// Game-specific messages are defined at game bridge level.
const (
	ControlMessageTypeCreateRoom    = "cre"
	ControlMessageTypeStartGame     = "ini"
	ControlMessageTypeJoinRoom      = "joi"
	ControlMessageTypeAddBot        = "bot"
	ControlMessageTypeKickPlayer    = "kck"
	ControlMessageTypePlayerQuits   = "qui"
	ControlMessageTypeTerminateRoom = "ter"
)

// MessageFromClient is a container struct used by
// clients to encapsulate their action messages sent to the hub.
type MessageFromClient struct {
	Author  Client
	Content MessageFromClientContent
}

// MessageFromClientContent is a struct that goes inside MessageFromClient struct, acting
// as a container for the different kind of parameters needed for each type
// of message
type MessageFromClientContent struct {
	Type   string          `json:"typ"`
	Params json.RawMessage `json:"par"`
}

// MessageCreateRoomParams defines the needed parameters for a create room
// message
type MessageCreateRoomParams struct {
	BridgeName string `json:"bri"`
}

// MessageJoinRoomParams defines the needed parameters for a join room message
type MessageJoinRoomParams struct {
	Room string `json:"rom"`
}

// MessageKickPlayerParams defines the needed parameters for a kick player
// message
type MessageKickPlayerParams struct {
	PlayerNumber int `json:"ply"`
}

// MessageAddBotParams defines the needed parameters for a add bot
// message
type MessageAddBotParams struct {
	BotLevel string `json:"lvl"`
}
