package api

import "encoding/json"

// AI is an interface that defines the minimum set of functions needed
// to implement an Artificial Intelligence
type AI interface {
	// FeedGameStatus updates the AI client with the current status of the game
	FeedGameStatus(json.RawMessage) error

	// Play makes the AI choose an action, returning it
	Play() Action
}
