package interfaces

type Client interface {
	ReadPump(channel interface{}, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Name() string
	SetName(v string) Client
	Close()
	IsBot() bool
	Room() Room
	SetRoom(r Room)
}
