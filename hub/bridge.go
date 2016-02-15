package hub

import (
	"encoding/json"
)

type Bridge interface {
	ParseMessage(t string, content json.RawMessage) ([]byte, error)
	CurrentPlayerNumber() int
	Status(n int) []byte
	AddPlayer() error
}
