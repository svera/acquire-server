package client

type MessageContent struct {
	Type   string      `json:"typ"`
	Params interface{} `json:"par"`
}

type Message struct {
	Author  *Client
	Content MessageContent
}

type PlayTileMessageParams struct {
	Tile string `json:"til"`
}

type NewCorpMessageParams struct {
	Corporation string `json:"cor"`
}

type BuyMessageParams struct {
	Corporations map[string]int `json:"cor"`
}

type SellTrade struct {
	Sell  int `json:"sel"`
	Trade int `json:"tra"`
}

type SellTradeMessageParams struct {
	Corporations map[string]SellTrade `json:"cor"`
}

type UntieMergeMessageParams struct {
	Corporation string `json:"cor"`
}
