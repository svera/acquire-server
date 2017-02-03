package interfaces

import "time"

// ControlMessageTypeStartGame defines the value that start game
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// When a game starts, an update message is broadcast to all clients with the initial status of the game.
// This update message format depends on the game, look at the corresponding game bridge documentation for details.
//
// The following is a StartGame message example:
//   {
//     "typ": "ini",
//     "par": {
//       "pto": 15
//     }
//   }
const ControlMessageTypeStartGame = "ini"

// ControlMessageTypeAddBot defines the value that add bot
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// A MessageCurrentPlayers is sent to all clients in the room when the bot is added.
//
// The following is an AddBot message example:
//   {
//     "typ": "bot",
//     "par": {
//       "lvl": "chaotic"
//     }
//   }
const ControlMessageTypeAddBot = "bot"

// ControlMessageTypeKickPlayer defines the value that kick player
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// A MessageCurrentPlayers is sent to all clients in the room when player is kicked.
//
// The following is a KickPlayer message example:
//   {
//     "typ": "kck",
//     "par": {
//       "ply": 2
//     }
//   }
const ControlMessageTypeKickPlayer = "kck"

// ControlMessageTypePlayerQuits defines the value that quit room
// messages must have in the Type field.
//
// A MessageClientOut is sent to the client when he/she quits.
//
// The following is a PlayerQuits message example:
//   {
//     "typ": "qui",
//     "par": {} // No params needed
//   }
const ControlMessageTypePlayerQuits = "qui"

// ControlMessageTypeSetClientData defines the value that set client data
// messages must have in the Type field.
//
// The following is a SetClientData message example:
//   {
//     "typ": "scd",
//     "par": {
//       "nam": "Sergio"
//     }
//   }
const ControlMessageTypeSetClientData = "scd"

// MessageKickPlayerParams defines the needed parameters for a kick player
// message.
type MessageKickPlayerParams struct {
	PlayerNumber int `json:"ply"`
}

// MessageAddBotParams defines the needed parameters for a add bot
// message.
type MessageAddBotParams struct {
	BotLevel string `json:"lvl"`
}

// MessageStartGameParams defines the needed parameters for a start game
// message.
type MessageStartGameParams struct {
	PlayerTimeout time.Duration `json:"pto"`
}

// MessageSetClientDataParams defines the needed parameters for a set client data
// message.
type MessageSetClientDataParams struct {
	Name string `json:"nam"`
}
