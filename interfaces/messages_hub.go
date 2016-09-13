package interfaces

// Control messages sent from the hub to the different players.
// These messages are common to all games.

const (
	ReasonRoomDestroyedTimeout    = "tim"
	ReasonRoomDestroyedTerminated = "ter"
	ReasonRoomDestroyedNoClients  = "ncl"
)

type MessageRoomCreated struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}

type MessageRoomDestroyed struct {
	Type   string `json:"typ"`
	Reason string `json:"rea"`
}

type MessageRoomsList struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}
