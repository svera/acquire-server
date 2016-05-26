package interfaces

// These constants define the different websocket close codes that signal
// the reason why that websocket connection was closed. Note that, following
// WS standard, all codes goes from 4000 onwards
// https://developer.mozilla.org/es/docs/Web/API/CloseEvent
const (
	HubDestroyed = 4000
	PlayerQuit   = 4001
	PlayerKicked = 4002
	PlayerLeft   = 4003
)

type Client interface {
	ReadPump(channel interface{}, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Owner() bool
	SetOwner(v bool) Client
	Name() string
	SetName(v string) Client
	Close(code int)
	IsBot() bool
}
