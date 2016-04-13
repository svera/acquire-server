package acquirebridge

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/svera/acquire"
	"github.com/svera/acquire/bots"
	"github.com/svera/acquire/corporation"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/player"
	"github.com/svera/acquire/tile"
	"github.com/svera/tbg-server/client"
	serverInterfaces "github.com/svera/tbg-server/interfaces"
)

// AcquireBridge implements the bridge interface in order to be able to have
// and acquire game through the turn based game server
type AcquireBridge struct {
	game    *acquire.Game
	players []acquireInterfaces.Player
}

const (
	// NotEndGame defines the message returned when a player claims wrongly that end game conditions have been met
	NotEndGame     = "not_end_game"
	minimumPlayers = 3
	maximumPlayers = 6
	// WrongMessage defines the message returned when AcquireBridge receives a malformed message
	WrongMessage = "message_parsing_error"
	// GameAlreadyStarted is an error returned when a player tries to start a game in a hub instance which an already running one
	GameAlreadyStarted = "game_already_started"
	// GameNotStarted is an error returned when a player tries to do an action that requires a running game
	GameNotStarted = "game_not_started"
	// GameFull is an error returned when a game already has the maximum number of players
	GameFull = "game_full"
	// InexistentPlayer is an error returned when someone tries to get information of a non existent player
	InexistentPlayer    = "inexistent_player"
	CorporationNotFound = "corporation_not_found"
)

// New initializes a new AcquireBridge instance
func New() *AcquireBridge {
	return &AcquireBridge{}
}

// ParseMessage gets an input JSON-encoded message and parses it, executing
// whatever actions are required by it
func (b *AcquireBridge) ParseMessage(t string, params json.RawMessage) ([]byte, error) {
	var err error
	var response []byte

	switch t {
	case messageTypePlayTile:
		var parsed playTileMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.playTile(parsed)
		}
	case messageTypeFoundCorporation:
		var parsed newCorpMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.foundCorporation(parsed)
		}
	case messageTypeBuyStock:
		var parsed buyMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.buyStock(parsed)
		}
	case messageTypeSellTrade:
		var parsed sellTradeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.sellTrade(parsed)
		}
	case messageTypeUntieMerge:
		var parsed untieMergeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.untieMerge(parsed)
		}
	case messageTypeEndGame:
		err = b.claimEndGame()
	default:
		err = errors.New(WrongMessage)
	}

	if err != nil {
		res := &errorMessage{
			Type:    "err",
			Content: err.Error(),
		}
		response, _ = json.Marshal(res)
	}
	return response, err
}

func (b *AcquireBridge) playTile(params playTileMessageParams) error {
	var err error
	var tl acquireInterfaces.Tile

	if tl, err = coordsToTile(params.Tile); err == nil {
		if err = b.game.PlayTile(tl); err == nil {
			return nil
		}
	}

	return err
}

func (b *AcquireBridge) foundCorporation(params newCorpMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	if corp, err = b.findCorpByName(params.Corporation); err == nil {
		if err = b.game.FoundCorporation(corp); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) buyStock(params buyMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	buy := map[acquireInterfaces.Corporation]int{}

	for corpName, amount := range params.Corporations {
		if corp, err = b.findCorpByName(corpName); err == nil {
			buy[corp] = amount
		} else {
			return err
		}
	}

	if err = b.game.BuyStock(buy); err == nil {
		return nil
	}
	return err
}

func (b *AcquireBridge) sellTrade(params sellTradeMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	sell := map[acquireInterfaces.Corporation]int{}
	trade := map[acquireInterfaces.Corporation]int{}

	for corpName, operation := range params.Corporations {
		if corp, err = b.findCorpByName(corpName); err == nil {
			sell[corp] = operation.Sell
			trade[corp] = operation.Trade
		} else {
			return err
		}
	}

	if err = b.game.SellTrade(sell, trade); err == nil {
		return nil
	}
	return err
}

func (b *AcquireBridge) untieMerge(params untieMergeMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	if corp, err = b.findCorpByName(params.Corporation); err == nil {
		if err = b.game.UntieMerge(corp); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) claimEndGame() error {
	if !b.game.ClaimEndGame().IsLastTurn() {
		return errors.New(NotEndGame)
	}
	return nil
}

func corpNames(corps []acquireInterfaces.Corporation) []string {
	names := []string{}
	for _, corp := range corps {
		names = append(names, corp.Name())
	}
	return names
}

func (b *AcquireBridge) findCorpByName(name string) (acquireInterfaces.Corporation, error) {
	for _, corp := range b.game.Corporations() {
		if strings.ToLower(corp.Name()) == strings.ToLower(name) {
			return corp, nil
		}
	}
	return &corporation.Corporation{}, errors.New(CorporationNotFound)
}

func (b *AcquireBridge) boardOwnership() map[string]string {
	cells := make(map[string]string)
	var letters = [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			cell := b.game.Board().Cell(number, letter)
			if cell.Type() == "corporation" {
				cells[strconv.Itoa(number)+letter] = strings.ToLower(cell.(*corporation.Corporation).Name())
			} else {
				cells[strconv.Itoa(number)+letter] = cell.Type()
			}
		}
	}

	return cells
}

func coordsToTile(tl string) (acquireInterfaces.Tile, error) {
	if len(tl) < 2 {
		return &tile.Tile{}, errors.New("Not a valid tile")
	}
	number, _ := strconv.Atoi(tl[:len(tl)-1])
	letter := string(tl[len(tl)-1 : len(tl)])
	return tile.New(number, letter), nil
}

// CurrentPlayerNumber returns the number of the player currently in turn
func (b *AcquireBridge) CurrentPlayerNumber() (int, error) {
	if !b.gameStarted() {
		return 0, errors.New(GameNotStarted)
	}
	return b.game.CurrentPlayerNumber(), nil
}

// gameStarted returns true if there's a game in progress, false otherwise
func (b *AcquireBridge) gameStarted() bool {
	if b.game == nil {
		return false
	}
	return true
}

// Status return a JSON string with the current status of the game
func (b *AcquireBridge) Status(n int) ([]byte, error) {
	playerInfo, rivalsInfo, err := b.playersInfo(n)
	if err != nil {
		return json.RawMessage{}, err
	}
	msg := statusMessage{
		Type:       "upd",
		Board:      b.boardOwnership(),
		State:      b.game.GameStateName(),
		Corps:      b.corpsData(),
		TiedCorps:  corpNames(b.game.TiedCorps()),
		Hand:       b.tilesData(b.players[n]),
		PlayerInfo: playerInfo,
		RivalsInfo: rivalsInfo,
		TurnNumber: b.game.Turn(),
		LastTurn:   b.game.IsLastTurn(),
	}
	response, _ := json.Marshal(msg)
	return response, err
}

func (b *AcquireBridge) tilesData(pl acquireInterfaces.Player) []handData {
	var hnd []handData
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, handData{
			Coords:   strconv.Itoa(tl.Number()) + tl.Letter(),
			Playable: b.game.IsTilePlayable(tl),
		})
	}
	return hnd
}

func (b *AcquireBridge) corpsData() [7]corpData {
	var data [7]corpData
	for i, corp := range b.game.Corporations() {
		data[i] = corpData{
			Name:            corp.Name(),
			Price:           corp.StockPrice(),
			MajorityBonus:   corp.MajorityBonus(),
			MinorityBonus:   corp.MinorityBonus(),
			RemainingShares: corp.Stock(),
			Size:            corp.Size(),
			Defunct:         b.game.IsCorporationDefunct(corp),
		}
	}
	return data
}

func (b *AcquireBridge) playersInfo(n int) (playerData, []playerData, error) {
	rivals := []playerData{}
	var ply playerData
	var err error

	if n < 0 || n >= len(b.players) {
		err = errors.New(InexistentPlayer)
	}
	for i, p := range b.players {
		if n != i {
			rivals = append(rivals, playerData{
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(i),
				Enabled:     b.isCurrentPlayer(i),
			})
		} else {
			ply = playerData{
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(n),
				Enabled:     b.isCurrentPlayer(n),
			}
		}
	}
	return ply, rivals, err
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

// AddPlayer adds a new player to the game
func (b *AcquireBridge) AddPlayer() error {
	if len(b.players) == maximumPlayers {
		return errors.New(GameFull)
	}
	if b.gameStarted() {
		return errors.New(GameAlreadyStarted)
	}
	b.players = append(b.players, player.New())
	return nil
}

// StartGame starts a new Acquire game
func (b *AcquireBridge) StartGame() error {
	var err error
	if b.gameStarted() {
		err = errors.New(GameAlreadyStarted)
	}
	b.game, err = acquire.New(b.players, acquire.Optional{})
	return err
}

func (b *AcquireBridge) IsGameOver() bool {
	if b.gameStarted() {
		return b.game.GameStateName() == acquireInterfaces.EndGameStateName
	}
	return false
}

func (b *AcquireBridge) AddBot(params interface{}) (serverInterfaces.Client, error) {
	if name, ok := params.(string); ok {
		if bot, err := bots.Create(name); err == nil {
			return NewBotClient(bot), nil
		} else {
			return &client.NullClient{}, err
		}
	}
	panic("Expecting string in AddBot parameter")
}
