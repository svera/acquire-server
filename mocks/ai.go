package mocks

import (
	"encoding/json"
)

// AI is a structure that implements the AI interface for testing
type AI struct {
	FakeIsInTurn    bool
	FakeMessageType string
	FakeMessage     json.RawMessage
	Calls           map[string]int
}

// FeedGameStatus mocks the FeedGameStatus method defined in the AI interface
func (a *AI) FeedGameStatus(json.RawMessage) error {
	a.Calls["FeedGameStatus"]++
	return nil

}

// IsInTurn mocks the IsInTurn method defined in the AI interface
func (a *AI) IsInTurn() bool {
	return a.FakeIsInTurn
}

// Play mocks the Play method defined in the AI interface
func (a *AI) Play() (string, json.RawMessage) {
	return a.FakeMessageType, a.FakeMessage
}
