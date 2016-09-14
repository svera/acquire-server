package interfaces

// Control messages sent from the room to the different players.
// These messages are common to all games.

// MessageSetOwner defines the needed parameters for a set owner
// message
type MessageSetOwner struct {
	Type string `json:"typ"`
	Role string `json:"rol"`
}

// MessageCurrentPlayers defines the needed parameters for a current players
// message
type MessageCurrentPlayers struct {
	Type   string          `json:"typ"`
	Values []MessagePlayer `json:"val"`
}

// MessagePlayer is a struct used inside MessageCurrentPlayers with data of a specific
// user
type MessagePlayer struct {
	Name  string `json:"nam"`
	Owner bool   `json:"own"`
}

// MessageError is a struct sent to a specific player
// when he/she does an action that leads to an error.
type MessageError struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

// MessageJoinRoomAccepted defines the needed parameters for a join room accepted
// message
type MessageJoinRoomAccepted struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}
