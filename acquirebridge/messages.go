package acquirebridge

const (
	messageTypePlayTile         = "ply"
	messageTypeFoundCorporation = "ncp"
	messageTypeBuyStock         = "buy"
	messageTypeSellTrade        = "sel"
	messageTypeUntieMerge       = "unt"
	messageTypeEndGame          = "end"
)

type playTileMessageParams struct {
	Tile string `json:"til"`
}

type newCorpMessageParams struct {
	Corporation string `json:"cor"`
}

type buyMessageParams struct {
	Corporations map[string]int `json:"cor"`
}

type sellTrade struct {
	Sell  int `json:"sel"`
	Trade int `json:"tra"`
}

type sellTradeMessageParams struct {
	Corporations map[string]sellTrade `json:"cor"`
}

type untieMergeMessageParams struct {
	Corporation string `json:"cor"`
}

type errorMessage struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

type handData struct {
	Coords   string `json:"coo"`
	Playable bool   `json:"pyb"`
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

type statusMessage struct {
	Type       string                `json:"typ"`
	Board      map[string]string     `json:"brd"`
	State      string                `json:"sta"`
	Corps      []corpData            `json:"cor"`
	TiedCorps  []string              `json:"tie"`
	PlayerInfo playerData            `json:"ply"`
	RivalsInfo map[string]playerData `json:"riv"`
	LastTurn   bool                  `json:"lst"`
}
