package interfaces

// Control messages sent from the hub to the different players.
// These messages are common to all games.

// Possible reasons why a room can be destroyed
const (
	ReasonRoomDestroyedTimeout    = "tim"
	ReasonRoomDestroyedTerminated = "ter"
	ReasonRoomDestroyedNoClients  = "ncl"
)

// MessageRoomCreated defines the needed parameters for a room created
// message
type MessageRoomCreated struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}

// MessageRoomDestroyed defines the needed parameters for a room destroyed
// message
type MessageRoomDestroyed struct {
	Type   string `json:"typ"`
	Reason string `json:"rea"`
}

// MessageRoomsList defines the needed parameters for a rooms list
// message
type MessageRoomsList struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}
