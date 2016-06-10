package hub

// Control messages sent from the hub to the different players.
// These messages are common to all games.

type newRoomCreatedMessage struct {
	Type string `json:"typ"`
	ID   string `json:"id"`
}

type joinRoomAcceptedMessage struct {
	Type string `json:"ack"`
	ID   string `json:"id"`
}

type setOwnerMessage struct {
	Type string `json:"typ"`
	Role string `json:"rol"`
}

type currentPlayersMessage struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}

// errorMessage is a struct sent to a specific player
// when he/she does an action that leads to an error.
type errorMessage struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}
