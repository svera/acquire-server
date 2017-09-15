package events

// Events emitted, always in past tense
const (
	GameStarted        = "gameStarted"
	ClientRegistered   = "clientRegistered"
	ClientUnregistered = "clientUnregistered"
	ClientOut          = "clientOut"
	ClientJoined       = "clientJoined"
	ClientsUpdated     = "clientsUpdated"
	GameStatusUpdated  = "gameStatusUpdated"
	RoomCreated        = "roomCreated"
	RoomDestroyed      = "roomDestroyed"
	Error              = "error"
)
