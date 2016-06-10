package interfaces

type Room interface {
	AddClient(c Client) error
	AddBot(level string) (Client, error)
	Owner() Client
	GameStarted() bool
	NumberClients() int
	ParseMessage(m *ClientMessage) error
	StartGame() error
	IsGameOver() bool
	Status(n int) ([]byte, error)
}
