package interfaces

// Control messages sent from the room to the different players.
// These messages are common to all games.

type MessageSetOwner struct {
	Type string `json:"typ"`
	Role string `json:"rol"`
}

type MessageCurrentPlayers struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}

// MessageError is a struct sent to a specific player
// when he/she does an action that leads to an error.
type MessageError struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

type MessageJoinRoomAccepted struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}
