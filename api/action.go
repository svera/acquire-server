package api

import "encoding/json"

// Action represents an action done by a player
type Action struct {
	PlayerName string

	// Type contains which action type this is
	Type string

	// Params contains additional values related to the action
	Params json.RawMessage
}
