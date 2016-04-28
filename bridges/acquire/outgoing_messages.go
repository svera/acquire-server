// This file specifies messages sent from the hub to the clients, basically notifying
// about the status of the game after a player action
package acquirebridge

// statusMessage is a struct which contains the status of the game at the moment
// it is issued. It is sent to each player after every action made by one of them.
type statusMessage struct {
	Type       string            `json:"typ"`
	Board      map[string]string `json:"brd"`
	State      string            `json:"sta"`
	Hand       []handData        `json:"hnd"`
	Corps      [7]corpData       `json:"cor"`
	TiedCorps  []string          `json:"tie"`
	PlayerInfo playerData        `json:"ply"`
	RivalsInfo []playerData      `json:"riv"`
	TurnNumber int               `json:"trn"`
	LastTurn   bool              `json:"lst"`
	History    []string          `json:"his"`
}

type corpData struct {
	Name            string `json:"nam"`
	Price           int    `json:"prc"`
	MajorityBonus   int    `json:"maj"`
	MinorityBonus   int    `json:"min"`
	RemainingShares int    `json:"rem"`
	Size            int    `json:"siz"`
	Defunct         bool   `json:"def"`
}

type playerData struct {
	Name        string `json:"nam"`
	Enabled     bool   `json:"ebl"`
	Cash        int    `json:"csh"`
	OwnedShares [7]int `json:"own"`
}

type handData struct {
	Coords   string `json:"coo"`
	Playable bool   `json:"pyb"`
}
