package interfaces

import "encoding/json"

// AI is an interface that defines the minimum set of functions needed
// to implement an Artificial Intelligence
type AI interface {
	FeedGameStatus(json.RawMessage) error
	Play() (string, json.RawMessage)
}
