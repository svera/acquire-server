package hub

type ErrorMessage struct {
	Type    string `json:"typ"`
	Content error  `json:"cnt"`
}

type CommonMessage struct {
	Type   string            `json:"typ"`
	Board  map[string]string `json:"brd"`
	Prices map[string]int    `json:"prc"`
}

type DirectMessage struct {
	CommonMessage
	Hand   []string       `json:"hnd"`
	Shares map[string]int `json:"shr"`
}
