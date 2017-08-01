package interfaces

import "encoding/json"

// Control messages sent to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed or why a player leaves a room.
// Used in MessageClientOut messages.
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
	TypeMessageClientOut        = "out"
	TypeMessageRoomsList        = "rms"
	TypeMessageCurrentPlayers   = "pls"
	TypeMessageError            = "err"
	TypeMessageJoinedRoom       = "joi"
	TypeMessageUpdateGameStatus = "upd"
)

// OutgoingMessage is a container struct used by
// the hub to encapsulate the messages sent to clients, adding common fields
// such as ID.
// The actual message coming from the backend is in Content.
type OutgoingMessage struct {
	ID   string
	Type string `json:"typ"`
	// SequenceNumber field that will allow clients to process incoming messages in order,
	// as they are not guaranteed to arrive at the same order they were sent
	// (for example, for update messages).
	SequenceNumber int             `json:"seq"`
	Content        json.RawMessage `json:"cnt"`
}

// MessageClientOut is sent to a client when he/she is expelled from a room.
// The following is a MessageClientOut message example:
//   {
//     "id": "550e8400-e29b-41d4-a716-446655440000",
//     "typ": "out",
//     "seq": 0,
//     "cnt": {
//       "rea": "tim"
//     }
//   }
type MessageClientOut struct {
	Reason string `json:"rea"`
}

// MessageRoomsList is sent to all clients when a new room is created.
// It contains all available rooms (rooms which haven't started a game yet).
// The following is a MessageRoomsList message example:
//   {
//     "id": "550e8400-e29b-41d4-a716-446655440000",
//     "typ": "rms",
//     "seq": 0,
//     "cnt": {
//       "val": ["VWXYZ", "ABCDE"]
//     }
//   }
type MessageRoomsList struct {
	Values []string `json:"val"`
}

// MessageCurrentPlayers is sent to all clients in a room when a player enters or leaves the room.
// The following is a MessageCurrentPlayers message example:
//   {
//     "id": "550e8400-e29b-41d4-a716-446655440000",
//     "typ": "pls",
//     "seq": 0,
//     "cnt": {
//       "val":
//       { // Indexed by player number
//	       "0": {"nam": "Miguel"},
//         "1": {"nam": "Sergio"}
//       }
//     }
//   }
type MessageCurrentPlayers struct {
	Values map[string]PlayerData `json:"val"`
}

// PlayerData is a struct used inside MessageCurrentPlayers with data of a specific
// user
type PlayerData struct {
	Name string `json:"nam"`
}

// MessageError is a message sent to a specific player
// when he/she does an action that leads to an error.
// The following is a MessageError message example:
//   {
//     "id": "550e8400-e29b-41d4-a716-446655440000",
//     "typ": "err",
//     "seq": 0,
//     "cnt": {
//       "cnt": "Whatever"
//     }
//   }
type MessageError struct {
	Content string `json:"cnt"`
}

// MessageJoinedRoom is a struct sent to a specific player
// when he/she joins to a room.
// The following is a MessageJoinedRoom message example:
//   {
//     "id": "550e8400-e29b-41d4-a716-446655440000",
//     "typ": "joi",
//     "seq": 0,
//     "cnt": {
//       "num": 2,
//       "id": "VWXYZ",
//       "own": false
//     }
//   }
type MessageJoinedRoom struct {
	ClientNumber int    `json:"num"`
	ID           string `json:"id"`
	// Owner signals if this client is the owner of the room
	Owner bool `json:"own"`
}
