package interfaces

import "encoding/json"

// IncomingMessage is a wrapper used by
// clients to encapsulate their action messages sent to the hub, adding metadata.
// The actual message coming from the frontend is in Content, Author
// field is set by the client itself.
type IncomingMessage struct {
	// Author is fulfilled automatically by the system whenever a message is received
	Author  Client
	Type    string          `json:"typ"`
	Content json.RawMessage `json:"cnt"`
}

// OutgoingMessage is a wrapper used by
// the hub to encapsulate the messages sent to clients, adding metadata.
// The actual message coming from the backend is in Content.
type OutgoingMessage struct {
	Type string `json:"typ"`
	// SequenceNumber field that will allow clients to process incoming messages in order,
	// as they are not guaranteed to arrive at the same order they were sent
	// (for example, for update messages).
	SequenceNumber int             `json:"seq,omitempty"`
	Content        json.RawMessage `json:"cnt"`
}
