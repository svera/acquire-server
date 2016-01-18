package hub

type Message struct {
	Result string            `json:"res"`
	Type   string            `json:"typ"`
	Board  map[string]string `json:"brd"`
	Hand   []string          `json:"hnd"`
}
