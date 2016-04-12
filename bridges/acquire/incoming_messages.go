// This file specifies messages sent from the clients to the hub, specifying an
// action
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
