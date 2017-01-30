package interfaces

import "encoding/json"

// IncomingMessage is a container struct used by
// clients to encapsulate their action messages sent to the hub.
// The actual message coming from the frontend is in Content, Author
// field is set by the client itself.
type IncomingMessage struct {
	Author  Client
	Content IncomingMessageContent
}

// IncomingMessageContent is a struct that goes inside IncomingMessage struct, acting
// as a container for the different kind of parameters needed for each type
// of message.
type IncomingMessageContent struct {
	Type   string          `json:"typ"`
	Params json.RawMessage `json:"par"`
}
