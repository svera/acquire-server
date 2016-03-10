package client

import (
	"encoding/json"
	"github.com/svera/tbg-server/interfaces"
)

type MessageContent struct {
	Type   string          `json:"typ"`
	Params json.RawMessage `json:"par"`
}

type Message struct {
	Author  interfaces.Client
	Content MessageContent
}
