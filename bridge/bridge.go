package bridge

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

type AcquireBridge struct {
	game    *acquire.Game
	players []interfaces.Player
}

const (
	// GameFull is an error returned when a game already has the maximum number of players
	GameFull       = "wrong_corporation_class"
	GameNotStarted = "game_not_started"
)

func (b *AcquireBridge) ParseMessage(t string, params json.RawMessage) ([]byte, error) {
	var err error
	var response []byte

	switch t {
	case "ini":
		b.game, err = acquire.New(
			board.New(),
			b.players,
			createCorporations(),
			tileset.New(),
			&fsm.PlayTile{},
		)

		//b.NewGameMergeTest()
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
		err = errors.New("Message parsing error")
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

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	if tl, err := coordsToTile(params.Tile); err == nil {
		if err := b.game.PlayTile(tl); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) foundCorporation(params newCorpMessageParams) error {
	var err error

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	if corp, err := b.findCorpByName(params.Corporation); err == nil {
		if err := b.game.FoundCorporation(corp); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) buyStock(params buyMessageParams) error {
	var err error

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	buy := map[interfaces.Corporation]int{}

	for corpName, amount := range params.Corporations {
		if corp, err := b.findCorpByName(corpName); err == nil {
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

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	sell := map[interfaces.Corporation]int{}
	trade := map[interfaces.Corporation]int{}

	for corpName, operation := range params.Corporations {
		if corp, err := b.findCorpByName(corpName); err == nil {
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

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	if corp, err := b.findCorpByName(params.Corporation); err == nil {
		if err := b.game.UntieMerge(corp); err == nil {
			return nil
		}
	}
	return err
}

func (b *AcquireBridge) claimEndGame() error {
	var err error

	if !b.gameStarted() {
		return errors.New(GameNotStarted)
	}

	if err := b.game.ClaimEndGame(); err == nil {
		return nil
	}
	return err
}

func (b *AcquireBridge) tilesToSlice(pl interfaces.Player) []string {
	var hnd []string
	for _, tl := range pl.Tiles() {
		hnd = append(hnd, strconv.Itoa(tl.Number())+tl.Letter())
	}
	return hnd
}

func corpNames(corps []interfaces.Corporation) []string {
	names := []string{}
	for _, corp := range corps {
		names = append(names, corp.Name())
	}
	return names
}

func (b *AcquireBridge) mapShares(pl interfaces.Player) map[string]int {
	corps := map[string]int{}
	for _, c := range b.game.DefunctCorporations() {
		if amount := pl.Shares(c); amount > 0 {
			corps[strings.ToLower(c.Name())] = amount
		}
	}
	return corps
}

func (b *AcquireBridge) findCorpByName(name string) (interfaces.Corporation, error) {
	for _, corp := range b.game.Corporations() {
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

func (b *AcquireBridge) CurrentPlayerNumber() int {
	if !b.gameStarted() {
		return 0
	}
	return b.game.CurrentPlayerNumber()
}

func (b *AcquireBridge) gameStarted() bool {
	if b.game == nil {
		return false
	}
	return true
}

func createCorporations() [7]interfaces.Corporation {
	var corps [7]interfaces.Corporation
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
				corps[i] = corp
			} else {
				panic(err)
			}
		}
	}
	return corps
}

func (b *AcquireBridge) Status(n int) []byte {
	pl := b.players[n]
	msg := statusMessage{
		Type:          "upd",
		Board:         b.boardOwnership(),
		Hand:          b.tilesToSlice(pl),
		State:         b.game.GameStateName(),
		InactiveCorps: corpNames(b.game.InactiveCorporations()),
		ActiveCorps:   corpNames(b.game.ActiveCorporations()),
		TiedCorps:     corpNames(b.game.TiedCorps()),
		Shares:        b.mapShares(pl),
		Enabled:       false,
	}
	if b.CurrentPlayerNumber() == n {
		msg.Enabled = true
	}
	response, _ := json.Marshal(msg)
	return response
}

func (b *AcquireBridge) AddPlayer() error {
	if len(b.players) == 6 {
		return errors.New(GameFull)
	}
	b.players = append(b.players, player.New())
	return nil
}
