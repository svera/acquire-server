package acquirebridge

import (
	"encoding/json"
	"errors"

	"github.com/svera/acquire"
	"github.com/svera/acquire/bots"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/sackson-server/bridges/acquire/corporation"
	"github.com/svera/sackson-server/bridges/acquire/player"
	serverInterfaces "github.com/svera/sackson-server/interfaces"
)

// AcquireBridge implements the bridge interface in order to be able to have
// and acquire game through the turn based game server
type AcquireBridge struct {
	game         *acquire.Game
	players      []acquireInterfaces.Player
	corporations [7]acquireInterfaces.Corporation
	history      []i18n
}

// NotEndGame defines the message returned when a player claims wrongly that end game conditions have been met
const NotEndGame = "not_end_game"

// WrongMessage defines the message returned when AcquireBridge receives a malformed message
const WrongMessage = "message_parsing_error"

// GameAlreadyStarted is an error returned when a player tries to start a game in a hub instance which an already running one
const GameAlreadyStarted = "game_already_started"

// GameNotStarted is an error returned when a player tries to do an action that requires a running game
const GameNotStarted = "game_not_started"

// InexistentPlayer is an error returned when someone tries to remove or get information of a non existent player
const InexistentPlayer = "inexistent_player"

// CorporationNotFound is an error returned when someone tries to use a non existent corporation
const CorporationNotFound = "corporation_not_found"

// New initializes a new AcquireBridge instance
func New() *AcquireBridge {
	return &AcquireBridge{
		corporations: defaultCorporations(),
	}
}

// Execute gets an input JSON-encoded message and parses it, executing
// whatever actions are required by it
func (b *AcquireBridge) Execute(clientName string, t string, params json.RawMessage) error {
	var err error
	b.history = nil

	switch t {
	case messageTypePlayTile:
		var parsed playTileMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.playTile(clientName, parsed)
		}
	case messageTypeFoundCorporation:
		var parsed newCorpMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.foundCorporation(clientName, parsed)
		}
	case messageTypeBuyStock:
		var parsed buyMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.buyStock(clientName, parsed)
		}
	case messageTypeSellTrade:
		var parsed sellTradeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.sellTrade(clientName, parsed)
		}
	case messageTypeUntieMerge:
		var parsed untieMergeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.untieMerge(clientName, parsed)
		}
	case messageTypeEndGame:
		err = b.claimEndGame(clientName)
	default:
		err = errors.New(WrongMessage)
	}

	return err
}

// CurrentPlayerNumber returns the number of the player currently in turn
func (b *AcquireBridge) CurrentPlayerNumber() (int, error) {
	if !b.GameStarted() {
		return 0, errors.New(GameNotStarted)
	}
	return b.game.CurrentPlayerNumber(), nil
}

// GameStarted returns true if there's a game in progress, false otherwise
func (b *AcquireBridge) GameStarted() bool {
	if b.game == nil {
		return false
	}
	return true
}

func (b *AcquireBridge) isCurrentPlayer(n int) bool {
	if number, err := b.CurrentPlayerNumber(); number == n && err == nil {
		return true
	}
	return false
}

func (b *AcquireBridge) playersShares(playerNumber int) [7]int {
	var data [7]int
	for i, corp := range b.game.Corporations() {
		data[i] = b.players[playerNumber].Shares(corp)
	}
	return data
}

// addPlayer adds a new player to the game
func (b *AcquireBridge) addPlayers(clients []serverInterfaces.Client) {
	for _, pl := range clients {
		b.players = append(b.players, player.New(pl.Name()))
	}
}

// DeactivatePlayer deactivates a player from the game
func (b *AcquireBridge) DeactivatePlayer(number int) error {
	if number < 0 || number > len(b.players) {
		return errors.New(InexistentPlayer)
	}
	playerName := b.players[number].(*player.Player).Name()
	b.game.DeactivatePlayer(b.players[number])
	b.history = append(b.history, i18n{
		Key: "game.history.player_left",
		Arguments: map[string]string{
			"player": playerName,
		},
	})
	return nil
}

// StartGame starts a new Acquire game
func (b *AcquireBridge) StartGame(clients []serverInterfaces.Client) error {
	var err error

	if b.GameStarted() {
		err = errors.New(GameAlreadyStarted)
	}

	b.addPlayers(clients)

	if b.game, err = acquire.New(b.players, acquire.Optional{Corporations: b.corporations}); err == nil {
		b.history = append(b.history, i18n{
			Key: "game.history.starter_player",
			Arguments: map[string]string{
				"player": b.currentPlayerName(),
			},
		})
	}
	return err
}

func (b *AcquireBridge) currentPlayerName() string {
	currentPlayerNumber, _ := b.CurrentPlayerNumber()
	return b.players[currentPlayerNumber].(*player.Player).Name()
}

// IsGameOver returns true if the game has reached its end or there are not
// enough players to continue playing
func (b *AcquireBridge) IsGameOver() bool {
	if b.GameStarted() {
		return b.game.GameStateName() == acquireInterfaces.EndGameStateName ||
			b.game.GameStateName() == acquireInterfaces.InsufficientPlayersStateName
	}
	return false
}

// AddBot adds a new bot
func (b *AcquireBridge) AddBot(params interface{}, room serverInterfaces.Room) (serverInterfaces.Client, error) {
	var err error
	var bot acquireInterfaces.Bot
	if level, ok := params.(string); ok {
		if bot, err = bots.Create(level); err == nil {
			return NewBotClient(bot, room), nil
		}
		return nil, err
	}
	panic("Expecting string in AddBot parameter")
}

func defaultCorporations() [7]acquireInterfaces.Corporation {
	var corporations [7]acquireInterfaces.Corporation
	corpsParams := [7]string{
		"Sackson",
		"Zeta",
		"Hydra",
		"Fusion",
		"America",
		"Phoenix",
		"Quantum",
	}

	for i, corpName := range corpsParams {
		corporations[i] = corporation.New(corpName, i)
	}
	return corporations
}
