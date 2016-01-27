package client

type MessageContent struct {
	Type   string                 `json:"typ"`
	Params map[string]interface{} `json:"par"`
}

type Message struct {
	Author  *Client
	Content MessageContent
}
