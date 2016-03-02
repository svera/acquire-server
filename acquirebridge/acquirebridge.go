package acquirebridge

import (
	"encoding/json"
	"errors"
	"github.com/svera/acquire"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/corporation"
	"github.com/svera/acquire/fsm"
	"github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/player"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
	"strconv"
	"strings"
)

// AcquireBridge implements the bridge interface in order to be able to have
// and acquire game through the turn based game server
type AcquireBridge struct {
	game         *acquire.Game
	players      []interfaces.Player
	corporations [7]interfaces.Corporation
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
)

// New initializes a new AcquireBridge instance
func New() *AcquireBridge {
	return &AcquireBridge{
		corporations: createCorporations(),
	}
}

// ParseMessage gets an input JSON-encoded message and parses it, executing
// whatever actions are required by it
func (b *AcquireBridge) ParseMessage(t string, params json.RawMessage) ([]byte, error) {
	var err error
	var response []byte

	switch t {
	case "ply":
		var parsed playTileMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.playTile(parsed)
		}
	case "ncp":
		var parsed newCorpMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.foundCorporation(parsed)
		}
	case "buy":
		var parsed buyMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.buyStock(parsed)
		}
	case "sel":
		var parsed sellTradeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.sellTrade(parsed)
		}
	case "unt":
		var parsed untieMergeMessageParams
		if err = json.Unmarshal(params, &parsed); err == nil {
			err = b.untieMerge(parsed)
		}
	case "end":
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
	var tl interfaces.Tile

	if tl, err = coordsToTile(params.Tile); err == nil {
		if err := b.game.PlayTile(tl); err == nil {
			return nil
		}
	}

	return err
}

func (b *AcquireBridge) foundCorporation(params newCorpMessageParams) error {
	var err error
	var corp interfaces.Corporation

	if corp, err = b.findCorpByName(params.Corporation); err == nil {
		if err = b.game.FoundCorporation(corp); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) buyStock(params buyMessageParams) error {
	var err error
	var corp interfaces.Corporation

	buy := map[interfaces.Corporation]int{}

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
	var corp interfaces.Corporation

	sell := map[interfaces.Corporation]int{}
	trade := map[interfaces.Corporation]int{}

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
	var corp interfaces.Corporation

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

func corpNames(corps []interfaces.Corporation) []string {
	names := []string{}
	for _, corp := range corps {
		names = append(names, corp.Name())
	}
	return names
}

func (b *AcquireBridge) findCorpByName(name string) (interfaces.Corporation, error) {
	for _, corp := range b.corporations {
		if strings.ToLower(corp.Name()) == name {
			return corp, nil
		}
	}
	return &corporation.Corporation{}, errors.New("corporation not found")
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

func coordsToTile(tl string) (interfaces.Tile, error) {
	if len(tl) < 2 {
		return &tile.Tile{}, errors.New("Not a valid tile")
	}
	number, _ := strconv.Atoi(tl[:len(tl)-1])
	letter := string(tl[len(tl)-1 : len(tl)])
	return tile.New(number, letter), nil
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

func createCorporations() [7]interfaces.Corporation {
	var corporations [7]interfaces.Corporation
	corpsParams := [7]map[string]int{
		map[string]int{"Sackson": 0},
		map[string]int{"Zeta": 0},
		map[string]int{"Hydra": 1},
		map[string]int{"Fusion": 1},
		map[string]int{"America": 1},
		map[string]int{"Phoenix": 2},
		map[string]int{"Quantum": 2},
	}

	for i, corpData := range corpsParams {
		for corpName, corpClass := range corpData {
			if corp, err := corporation.New(corpName, corpClass); err == nil {
				corporations[i] = corp
			} else {
				panic(err)
			}
		}
	}
	return corporations
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
		PlayerInfo: playerInfo,
		RivalsInfo: rivalsInfo,
		LastTurn:   b.game.IsLastTurn(),
	}
	response, _ := json.Marshal(msg)
	return response, err
}

func (b *AcquireBridge) tilesData(pl interfaces.Player) []handData {
	var hnd []handData
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, handData{
			Coords:   strconv.Itoa(tl.Number()) + tl.Letter(),
			Playable: b.game.IsTilePlayable(tl),
		})
	}
	return hnd
}

func (b *AcquireBridge) corpsData() []corpData {
	var data []corpData
	for _, corp := range b.corporations {
		data = append(data, corpData{
			Name:            corp.Name(),
			Price:           corp.StockPrice(),
			MajorityBonus:   corp.MajorityBonus(),
			MinorityBonus:   corp.MinorityBonus(),
			RemainingShares: corp.Stock(),
			Size:            corp.Size(),
			Defunct:         b.game.IsCorporationDefunct(corp),
		})
	}
	return data
}

func (b *AcquireBridge) playersInfo(n int) (playerData, map[string]playerData, error) {
	rivals := map[string]playerData{}
	var ply playerData
	var err error
	var number int
	for i, p := range b.players {
		if n != i {
			rivals[strconv.Itoa(i)] = playerData{
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(i),
			}
		} else {
			ply = playerData{
				Enabled:     false,
				Hand:        b.tilesData(p),
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(i),
			}
			if number, err = b.CurrentPlayerNumber(); number == n && err == nil {
				ply.Enabled = true
			}
		}
	}
	return ply, rivals, err
}

func (b *AcquireBridge) playersShares(playerNumber int) []int {
	var data []int
	for _, corp := range b.corporations {
		data = append(data, b.players[playerNumber].Shares(corp))
	}
	return data
}

// AddPlayer adds a new player to the game
func (b *AcquireBridge) AddPlayer() error {
	if len(b.players) == maximumPlayers {
		return errors.New(GameFull)
	}
	b.players = append(b.players, player.New())
	return nil
}

// StartGame starts a new Acquire game
func (b *AcquireBridge) StartGame() error {
	var err error
	if b.game != nil {
		err = errors.New(GameAlreadyStarted)
	}
	b.game, err = acquire.New(
		board.New(),
		b.players,
		b.corporations,
		tileset.New(),
		&fsm.PlayTile{},
	)
	return err
}
