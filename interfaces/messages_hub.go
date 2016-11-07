package interfaces

// Control messages sent from the hub to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed
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
	TypeMessageRoomCreated = "new"
	TypeMessageClientOut   = "out"
	TypeMessageRoomsList   = "rms"
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
