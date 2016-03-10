package hub

// Control messages sent from the hub to the different players.
// These messages are common to all games.

const (
	controlMessageTypeStartGame = "ini"
	controlMessageTypeAddBot    = "bot"
)

type setOwnerMessage struct {
	Type string `json:"typ"`
	Role string `json:"rol"`
}

type currentPlayersMessage struct {
	Type   string   `json:"typ"`
	Values []string `json:"val"`
}
