package hub

type ErrorMessage struct {
	Type    string `json:"typ"`
	Content string `json:"cnt"`
}

type CommonMessage struct {
	Type    string            `json:"typ"`
	Board   map[string]string `json:"brd"`
	Prices  map[string]int    `json:"prc"`
	Enabled bool              `json:"ebl"`
}

type DirectMessage struct {
	Type          string         `json:"typ"`
	Hand          []string       `json:"hnd"`
	Shares        map[string]int `json:"sha"`
	State         string         `json:"sta"`
	InactiveCorps []string       `json:"ina"`
	ActiveCorps   []string       `json:"act"`
}
