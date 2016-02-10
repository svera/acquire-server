package client

import "encoding/json"

type MessageContent struct {
	Type   string          `json:"typ"`
	Params json.RawMessage `json:"par"`
}

type Message struct {
	Author  *Client
	Content MessageContent
}
