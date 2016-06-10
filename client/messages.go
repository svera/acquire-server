package client

// Control messages types, common to all games.
// Game-specific messages are defined at game bridge level.
const (
	ControlMessageTypeCreateRoom    = "cre"
	ControlMessageTypeStartGame     = "ini"
	ControlMessageTypeJoinPlayer    = "joi"
	ControlMessageTypeAddBot        = "bot"
	ControlMessageTypeKickPlayer    = "kck"
	ControlMessageTypePlayerQuits   = "qui"
	ControlMessageTypeTerminateRoom = "ter"
)

// CreateRoomMessageParams contains the needed parameters for a create room
// message
type CreateRoomMessageParams struct {
	BridgeName string `json:"bri"`
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
