package messages

import (
	"encoding/json"
	"time"
)

// Control messages sent to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed or why a player leaves a room.
// Used in ClientOut messages.
const (
	ReasonRoomDestroyedTimeout      = "tim"
	ReasonRoomDestroyedTerminated   = "ter"
	ReasonRoomDestroyedNoClients    = "ncl"
	ReasonRoomDestroyedGamePanicked = "pan"
	ReasonPlayerTimedOut            = "ptm"
	ReasonPlayerKicked              = "kck"
	ReasonPlayerQuitted             = "qui"
)

// Types for the messages sent to the clients.
const (
	TypeClientOut        = "out"
	TypeRoomsList        = "rms"
	TypeCurrentPlayers   = "pls"
	TypeError            = "err"
	TypeJoinedRoom       = "joi"
	TypeUpdateGameStatus = "upd"
	TypeGameStarted      = "gst"
)

// ClientOut is sent to a client when he/she is expelled from a room.
// The following is a ClientOut message example:
//   {
//     "typ": "out",
//     "cnt": {
//       "rea": "tim"
//     }
//   }
type ClientOut struct {
	Reason string `json:"rea"`
}

// RoomsList is sent to all clients when a new room is created.
// It contains all available rooms (rooms which haven't started a game yet).
// The following is a RoomsList message example:
//   {
//     "typ": "rms",
//     "cnt": {
//       "val": ["VWXYZ", "ABCDE"]
//     }
//   }
type RoomsList struct {
	Values []string `json:"val"`
}

// CurrentPlayers is sent to all clients in a room when a player enters or leaves the room.
// The following is a CurrentPlayers message example:
//   {
//     "typ": "pls",
//     "cnt": {
//       "val":
//       { // Indexed by player number
//	       "0": {"nam": "Miguel"},
//         "1": {"nam": "Sergio"}
//       }
//     }
//   }
type CurrentPlayers struct {
	Values map[string]PlayerData `json:"val"`
}

// PlayerData is a struct used inside MessageCurrentPlayers with data of a specific
// user
type PlayerData struct {
	Name string `json:"nam"`
}

// Error is a message sent to a specific player
// when he/she does an action that leads to an error.
// The following is a Error message example:
//   {
//     "typ": "err",
//     "cnt": {
//       "des": "Whatever"
//     }
//   }
type Error struct {
	Description string `json:"des"`
}

// JoinedRoom is a message sent to a specific player
// when he/she joins to a room.
// The following is a JoinedRoom message example:
//   {
//     "typ": "joi",
//     "cnt": {
//       "num": 2,
//       "id": "VWXYZ",
//       "own": false
//     }
//   }
type JoinedRoom struct {
	ClientNumber int    `json:"num"`
	ID           string `json:"id"`
	// Owner signals if this client is the owner of the room
	Owner bool `json:"own"`
}

// GameStarted is a message sent to all players
// when a game starts.
// The following is a GameStarted message example:
//   {
//     "typ": "gst",
//     "cnt": {
//       "pto": 0,
//       "gpa": {
//         ···
//       }
//     }
//   }
type GameStarted struct {
	PlayerTimeOut  time.Duration   `json:"pto"`
	GameParameters json.RawMessage `json:"gpa"`
}
