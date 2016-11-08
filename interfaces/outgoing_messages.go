package interfaces

// Control messages sent to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed or why a player leaves a room
const (
	ReasonRoomDestroyedTimeout    = "tim"
	ReasonRoomDestroyedTerminated = "ter"
	ReasonRoomDestroyedNoClients  = "ncl"
	ReasonPlayerTimedOut          = "ptm"
	ReasonPlayerKicked            = "kck"
	ReasonPlayerQuitted           = "qui"
)

// Types for the messages in this file
const (
	TypeMessageRoomCreated    = "new"
	TypeMessageClientOut      = "out"
	TypeMessageRoomsList      = "rms"
	TypeMessageCurrentPlayers = "pls"
	TypeMessageError          = "err"
)

// MessageRoomCreated defines the needed parameters for a room created
// message
type MessageRoomCreated struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}

// MessageClientOut defines the needed parameters for a client out
// message
type MessageClientOut struct {
	Type   string `json:"typ"`
	Reason string `json:"rea"`
}

// MessageRoomsList defines the needed parameters for a rooms list
// message
type MessageRoomsList struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}

// MessageCurrentPlayers defines the needed parameters for a current players
// message
type MessageCurrentPlayers struct {
	Type   string       `json:"typ"`
	Values []PlayerData `json:"val"`
}

// PlayerData is a struct used inside MessageCurrentPlayers with data of a specific
// user
type PlayerData struct {
	Name  string `json:"nam"`
	Owner bool   `json:"own"`
}

// MessageError is a struct sent to a specific player
// when he/she does an action that leads to an error.
type MessageError struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}
