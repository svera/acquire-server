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

// TypeUpdateGameStatus defines the value that update game status
// messages must have in the Type field.
const TypeUpdateGameStatus = "upd"

// TypeClientOut defines the value that client out
// messages must have in the Type field.
//
// ClientOut is sent to a client when he/she is expelled from a room.
// The following is a ClientOut message example:
//   {
//     "typ": "out",
//     "cnt": {
//       "rea": "tim"
//     }
//   }
const TypeClientOut = "out"

// ClientOut defines the needed parameters for a client out
// message.
type ClientOut struct {
	Reason string `json:"rea"`
}

// TypeRoomsList defines the value that rooms list
// messages must have in the Type field.
//
// RoomsList is sent to all clients when a new room is created.
// It contains all available rooms (rooms which haven't started a game yet).
// The following is a RoomsList message example:
//   {
//     "typ": "rms",
//     "cnt": {
//       "val": ["VWXYZ", "ABCDE"]
//     }
//   }
const TypeRoomsList = "rms"

// RoomsList defines the needed parameters for a rooms list
// message.
type RoomsList struct {
	Values []string `json:"val"`
}

// TypeCurrentPlayers defines the value that current players
// messages must have in the Type field.
//
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
const TypeCurrentPlayers = "pls"

// CurrentPlayers defines the needed parameters for a current players
// message.
type CurrentPlayers struct {
	Values map[string]PlayerData `json:"val"`
}

// PlayerData is a struct used inside MessageCurrentPlayers with data of a specific
// user
type PlayerData struct {
	Name string `json:"nam"`
}

// TypeError defines the value that error
// messages must have in the Type field.
//
// Error is a message sent to a specific player
// when he/she does an action that leads to an error.
// The following is a Error message example:
//   {
//     "typ": "err",
//     "cnt": {
//       "des": "Whatever"
//     }
//   }
const TypeError = "err"

// Error defines the needed parameters for an error
// message.
type Error struct {
	Description string `json:"des"`
}

// TypeJoinedRoom defines the value that joined room
// messages must have in the Type field.
//
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
const TypeJoinedRoom = "joi"

// JoinedRoom defines the needed parameters for a joined room
// message.
type JoinedRoom struct {
	ClientNumber int    `json:"num"`
	ID           string `json:"id"`
	// Owner signals if this client is the owner of the room
	Owner bool `json:"own"`
}

// TypeGameStarted defines the value that game started
// messages must have in the Type field.
//
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
const TypeGameStarted = "gst"

// GameStarted defines the needed parameters for a game started
// message.
type GameStarted struct {
	PlayerTimeOut  time.Duration   `json:"pto"`
	GameParameters json.RawMessage `json:"gpa"`
}
