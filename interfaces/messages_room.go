package interfaces

// Control messages sent from the room to the different players.
// These messages are common to all games.

// Types for the messages in this file
const (
	TypeMessageCurrentPlayers = "pls"
	TypeMessageError          = "err"
)

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
