package interfaces

type Room interface {
	GameStarted() bool
	NumberClients() int
	ParseMessage(m *ClientMessage) (map[Client][]byte, error)
	IsGameOver() bool
	RemoveClient(c Client) map[Client][]byte
	ID() string
}
