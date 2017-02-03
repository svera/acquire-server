package interfaces

// Control messages sent to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed or why a player leaves a room.
// Used in MessageClientOut messages.
const (
	ReasonRoomDestroyedTimeout    = "tim"
	ReasonRoomDestroyedTerminated = "ter"
	ReasonRoomDestroyedNoClients  = "ncl"
	ReasonPlayerTimedOut          = "ptm"
	ReasonPlayerKicked            = "kck"
	ReasonPlayerQuitted           = "qui"
)

// Types for the messages sent to the clients.
const (
	TypeMessageClientOut      = "out"
	TypeMessageRoomsList      = "rms"
	TypeMessageCurrentPlayers = "pls"
	TypeMessageError          = "err"
	TypeMessageJoinedRoom     = "joi"
)

// MessageClientOut is sent to a client when he/she is expelled from a room.
// The following is a MessageClientOut message example:
//   {
//     "typ": "out",
//     "rea": "tim"
//   }
type MessageClientOut struct {
	Type   string `json:"typ"`
	Reason string `json:"rea"`
}

// MessageRoomsList is sent to all clients when a new room is created.
// It contains all available rooms (rooms which haven't started a game yet).
// The following is a MessageRoomsList message example:
//   {
//     "typ": "rms",
//     "val": ["VWXYZ", "ABCDE"]
//   }
type MessageRoomsList struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}

// MessageCurrentPlayers is sent to all clients in a room when a player enters or leaves the room.
// The following is a MessageCurrentPlayers message example:
//   {
//     "typ": "pls",
//     "val":
//     [
//       {"nam": "Miguel"},
//       {"nam": "Sergio"}
//     ]
//   }
type MessageCurrentPlayers struct {
	Type   string       `json:"typ"`
	Values []PlayerData `json:"val"`
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
//     "typ": "err",
//     "cnt": "Whatever"
//   }
type MessageError struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

// MessageJoinedRoom is a struct sent to a specific player
// when he/she joins to a room.
// The following is a MessageJoinedRoom message example:
//   {
//     "typ": "joi",
//     "num": 2,
//     "id": "VWXYZ",
//     "own": false
//   }
type MessageJoinedRoom struct {
	Type         string `json:"typ"`
	ClientNumber int    `json:"num"`
	ID           string `json:"id"`
	// Owner signals if this client is the owner of the room
	Owner bool `json:"own"`
}
