package drivers

import (
	"encoding/json"

	"github.com/svera/sackson-server/api"
)

// AI is a structure that implements the AI interface for testing
type AI struct {
	FakePlay api.Action
	Calls    map[string]int
}

// FeedGameStatus mocks the FeedGameStatus method defined in the AI interface
func (a *AI) FeedGameStatus(json.RawMessage) error {
	a.Calls["FeedGameStatus"]++
	return nil

}

// Play mocks the Play method defined in the AI interface
func (a *AI) Play() api.Action {
	return a.FakePlay
}
