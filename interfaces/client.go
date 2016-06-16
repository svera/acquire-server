package interfaces

type Client interface {
	ReadPump(channel interface{}, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Name() string
	SetName(v string) Client
	Close(code int)
	IsBot() bool
	Room() Room
	SetRoom(r Room)
}
