package interfaces

type Client interface {
	ReadPump(channel interface{}, unregister chan Client)
	WritePump()
	Incoming() chan []byte
	Owner() bool
	SetOwner(v bool) Client
	Name() string
	SetName(v string) Client
	Close()
}
