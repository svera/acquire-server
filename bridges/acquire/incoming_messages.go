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
	CorporationIndex int `json:"cor"`
}

type buyMessageParams struct {
	CorporationsIndexes map[string]int `json:"cor"`
}

type sellTradeMessageParams struct {
	CorporationsIndexes map[string]sellTrade `json:"cor"`
}

type sellTrade struct {
	Sell  int `json:"sel"`
	Trade int `json:"tra"`
}

type untieMergeMessageParams struct {
	CorporationIndex int `json:"cor"`
}
