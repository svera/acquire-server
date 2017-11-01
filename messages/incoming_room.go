package messages

import (
	"encoding/json"
	"time"
)

// TypeStartGame defines the value that start game
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// When a game starts, an update message is broadcast to all clients with the initial status of the game.
// This update message format depends on the game, look at the corresponding game driver documentation for details.
//
// The following is a StartGame message example:
//   {
//     "typ": "ini",
//     "cnt": {
//       "pto": 15,
//       "gpa": {
//         ···
//       }
//     }
//   }
const TypeStartGame = "ini"

// StartGame defines the needed parameters for a start game
// message.
type StartGame struct {
	PlayerTimeout  time.Duration   `json:"pto"`
	GameParameters json.RawMessage `json:"gpa"`
}

// TypeAddBot defines the value that add bot
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// A MessageCurrentPlayers is sent to all clients in the room when the bot is added.
//
// The following is an AddBot message example:
//   {
//     "typ": "bot",
//     "cnt": {
//       "lvl": "chaotic"
//     }
//   }
const TypeAddBot = "bot"

// AddBot defines the needed parameters for a add bot
// message.
type AddBot struct {
	BotLevel string `json:"lvl"`
}

// TypeKickPlayer defines the value that kick player
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// A MessageCurrentPlayers is sent to all clients in the room when player is kicked.
//
// The following is a KickPlayer message example:
//   {
//     "typ": "kck",
//     "cnt": {
//       "ply": 2
//     }
//   }
const TypeKickPlayer = "kck"

// KickPlayer defines the needed parameters for a kick player
// message.
type KickPlayer struct {
	PlayerNumber int `json:"ply"`
}

// TypePlayerQuits defines the value that quit room
// messages must have in the Type field.
//
// A MessageClientOut is sent to the client when he/she quits.
//
// The following is a PlayerQuits message example:
//   {
//     "typ": "qui",
//     "cnt": {} // No params needed
//   }
const TypePlayerQuits = "qui"

// TypeSetClientData defines the value that set client data
// messages must have in the Type field.
//
// The following is a SetClientData message example:
//   {
//     "typ": "scd",
//     "cnt": {
//       "nam": "Sergio"
//     }
//   }
const TypeSetClientData = "scd"

// SetClientDataParams defines the needed parameters for a set client data
// message.
type SetClientDataParams struct {
	Name string `json:"nam"`
}
