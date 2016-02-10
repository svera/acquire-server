package interfaces

import (
	"encoding/json"
)

type Bridge interface {
	ParseMessage(t string, content json.RawMessage) ([]byte, error)
	CurrentPlayerNumber() int
	NewGame()
	Status(n int) []byte
	AddPlayer()
	// Testing function, to be deleted
	NewGameMergeTest()
}
