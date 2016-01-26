package client

type MessageContent struct {
	Type   string            `json:"typ"`
	Params map[string]string `json:"det"`
}

type Message struct {
	Author  *Client
	Content MessageContent
}
