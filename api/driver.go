package api

// Driver is an interface that defines the minimum set of functions needed
// to implement a game driver which can be used within a hub instance
type Driver interface {
	// Execute gets an action made by clientName and executes it
	Execute(action Action) error

	// CurrentPlayersNumbers returns a slice containing the number of each player currently in turn
	CurrentPlayersNumbers() ([]int, error)

	// Status return a status message with the current status of the game,
	// with player specific information for the passed playerNumber
	Status(playerNumber int) (interface{}, error)

	// RemovePlayer removes the player with the passed number from the game
	RemovePlayer(number int) error

	// CreateAI creates and returns an instance of an AI with the passed parameters
	CreateAI(params interface{}) (AI, error)

	// StartGame starts a new game
	StartGame(clientNames map[int]string) error

	// GameStarted returns true if there's a game in progress, false otherwise
	GameStarted() bool

	// IsGameOver returns true if the game has reached its end or there are not
	// enough players to continue playing
	IsGameOver() bool

	// Name returns the name of the driver, used to identify which game it implements
	Name() string
}
