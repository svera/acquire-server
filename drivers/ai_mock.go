package drivers

import (
	"encoding/json"
)

// AI is a structure that implements the AI interface for testing
type AI struct {
	FakeMessageType string
	FakeMessage     json.RawMessage
	Calls           map[string]int
}

// FeedGameStatus mocks the FeedGameStatus method defined in the AI interface
func (a *AI) FeedGameStatus(json.RawMessage) error {
	a.Calls["FeedGameStatus"]++
	return nil

}

// Play mocks the Play method defined in the AI interface
func (a *AI) Play() (string, json.RawMessage) {
	return a.FakeMessageType, a.FakeMessage
}
