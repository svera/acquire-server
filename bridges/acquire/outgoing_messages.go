package acquirebridge

// This file specifies messages sent from the hub to the clients, basically notifying
// about the status of the game after a player action

// statusMessage is a struct which contains the status of the game at the moment
// it is issued. It is sent to each player after every action made by one of them.
type statusMessage struct {
	Type        string            `json:"typ"`
	Board       map[string]string `json:"brd"`
	State       string            `json:"sta"`
	Hand        map[string]bool   `json:"hnd"`
	Corps       [7]corpData       `json:"cor"`
	TiedCorps   []int             `json:"tie"`
	PlayerInfo  playerData        `json:"ply"`
	RivalsInfo  []playerData      `json:"riv"`
	RoundNumber int               `json:"rnd"`
	IsLastRound bool              `json:"lst"`
	History     []i18n            `json:"his"`
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
	Active      bool   `json:"atv"`
	Name        string `json:"nam"`
	InTurn      bool   `json:"trn"`
	Cash        int    `json:"csh"`
	OwnedShares [7]int `json:"own"`
}

type i18n struct {
	Key       string            `json:"key"`
	Arguments map[string]string `json:"arg"`
}
