package acquirebridge

const (
	messageTypePlayTile         = "ply"
	messageTypeFoundCorporation = "ncp"
	messageTypeBuyStock         = "buy"
	messageTypeSellTrade        = "sel"
	messageTypeUntieMerge       = "unt"
	messageTypeEndGame          = "end"
)

// playTileMessageParams is a struct which defines the format of the params of
// incoming playTile messages
type playTileMessageParams struct {
	Tile string `json:"til"`
}

type newCorpMessageParams struct {
	Corporation string `json:"cor"`
}

type buyMessageParams struct {
	Corporations map[string]int `json:"cor"`
}

type sellTradeMessageParams struct {
	Corporations map[string]sellTrade `json:"cor"`
}

type sellTrade struct {
	Sell  int `json:"sel"`
	Trade int `json:"tra"`
}

type untieMergeMessageParams struct {
	Corporation string `json:"cor"`
}

// errorMessage is a struct sent to an specific player
// when he/she does an action that leads to an error.
type errorMessage struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

// statusMessage is a struct which contains the status of the game at the moment
// it is issued. It is sent to each player after every action made by one of them.
type statusMessage struct {
	Type       string                `json:"typ"`
	Board      map[string]string     `json:"brd"`
	State      string                `json:"sta"`
	Corps      [7]corpData           `json:"cor"`
	TiedCorps  []string              `json:"tie"`
	PlayerInfo playerData            `json:"ply"`
	RivalsInfo map[string]playerData `json:"riv"`
	LastTurn   bool                  `json:"lst"`
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
	Enabled     bool       `json:"ebl"`
	Hand        []handData `json:"hnd"`
	Cash        int        `json:"csh"`
	OwnedShares []int      `json:"own"`
}

type handData struct {
	Coords   string `json:"coo"`
	Playable bool   `json:"pyb"`
}
