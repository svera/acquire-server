package bridge

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

type statusMessage struct {
	Type          string            `json:"typ"`
	Board         map[string]string `json:"brd"`
	Prices        map[string]int    `json:"prc"`
	Enabled       bool              `json:"ebl"`
	Hand          []string          `json:"hnd"`
	Shares        map[string]int    `json:"sha"`
	State         string            `json:"sta"`
	InactiveCorps []string          `json:"ina"`
	ActiveCorps   []string          `json:"act"`
	TiedCorps     []string          `json:"tie"`
}
