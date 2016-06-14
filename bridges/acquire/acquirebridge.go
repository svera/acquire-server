package acquirebridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/svera/acquire"
	"github.com/svera/acquire/bots"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/tile"
	"github.com/svera/tbg-server/bridges/acquire/corporation"
	"github.com/svera/tbg-server/bridges/acquire/player"
	serverInterfaces "github.com/svera/tbg-server/interfaces"
)

// AcquireBridge implements the bridge interface in order to be able to have
// and acquire game through the turn based game server
type AcquireBridge struct {
	game         *acquire.Game
	players      []acquireInterfaces.Player
	corporations [7]acquireInterfaces.Corporation
	history      []string
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
	// InexistentPlayer is an error returned when someone tries to remove or get information of a non existent player
	InexistentPlayer = "inexistent_player"
	// CorporationNotFound is an error returned when someone tries to use a non existent corporation
	CorporationNotFound = "corporation_not_found"
)

// New initializes a new AcquireBridge instance
func New() *AcquireBridge {
	return &AcquireBridge{
		corporations: defaultCorporations(),
	}
}

// Execute gets an input JSON-encoded message and parses it, executing
// whatever actions are required by it
func (b *AcquireBridge) Execute(t string, params json.RawMessage) error {
	var err error
	b.history = nil

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

	return err
}

func (b *AcquireBridge) playTile(params playTileMessageParams) error {
	var err error
	var tl acquireInterfaces.Tile

	if tl, err = coordsToTile(params.Tile); err == nil {
		if err = b.game.PlayTile(tl); err == nil {
			b.history = append(b.history, fmt.Sprintf("%s played tile %s", b.currentPlayerName(), params.Tile))
			return nil
		}
	}

	return err
}

func (b *AcquireBridge) foundCorporation(params newCorpMessageParams) error {
	if params.CorporationIndex < 0 || params.CorporationIndex > 6 {
		return errors.New(CorporationNotFound)
	}
	if err := b.game.FoundCorporation(b.corporations[params.CorporationIndex]); err != nil {
		return err
	}
	return nil
}

func (b *AcquireBridge) buyStock(params buyMessageParams) error {
	buy := map[acquireInterfaces.Corporation]int{}

	for corpIndex, amount := range params.CorporationsIndexes {
		index, _ := strconv.Atoi(corpIndex)
		if index < 0 || index > 6 {
			return errors.New(CorporationNotFound)
		}

		buy[b.corporations[index]] = amount
	}

	if err := b.game.BuyStock(buy); err != nil {
		return err
	}
	return nil
}

func (b *AcquireBridge) sellTrade(params sellTradeMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	sell := map[acquireInterfaces.Corporation]int{}
	trade := map[acquireInterfaces.Corporation]int{}

	for corpIndex, operation := range params.CorporationsIndexes {
		index, _ := strconv.Atoi(corpIndex)
		if index < 0 || index > 6 {
			return errors.New(CorporationNotFound)
		}
		corp = b.corporations[index]
		sell[corp] = operation.Sell
		trade[corp] = operation.Trade
	}

	if err = b.game.SellTrade(sell, trade); err != nil {
		return err
	}
	return nil
}

func (b *AcquireBridge) untieMerge(params untieMergeMessageParams) error {
	if params.CorporationIndex < 0 || params.CorporationIndex > 6 {
		return errors.New(CorporationNotFound)
	}

	if err := b.game.UntieMerge(b.corporations[params.CorporationIndex]); err != nil {
		return err
	}
	return nil
}

func (b *AcquireBridge) claimEndGame() error {
	if !b.game.ClaimEndGame().IsLastRound() {
		return errors.New(NotEndGame)
	}
	return nil
}

func corpIndexes(corps []acquireInterfaces.Corporation) []int {
	indexes := []int{}
	for _, corp := range corps {

		indexes = append(indexes, corp.(*corporation.Corporation).Index())
	}
	return indexes
}

func (b *AcquireBridge) boardOwnership() map[string]string {
	cells := make(map[string]string)
	var letters = [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			cell := b.game.Board().Cell(number, letter)
			if cell.Type() == "corporation" {
				cells[strconv.Itoa(number)+letter] = fmt.Sprintf("c%d", cell.(*corporation.Corporation).Index())
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

// Status return a JSON string with the current status of the game
func (b *AcquireBridge) Status(n int) ([]byte, error) {
	if !b.GameStarted() {
		return nil, errors.New(GameNotStarted)
	}

	playerInfo, rivalsInfo, err := b.playersInfo(n)
	if err != nil {
		return json.RawMessage{}, err
	}
	msg := statusMessage{
		Type:        "upd",
		Board:       b.boardOwnership(),
		State:       b.game.GameStateName(),
		Corps:       b.corpsData(),
		TiedCorps:   corpIndexes(b.game.TiedCorps()),
		Hand:        b.tilesData(b.players[n]),
		PlayerInfo:  playerInfo,
		RivalsInfo:  rivalsInfo,
		RoundNumber: b.game.Round(),
		IsLastRound: b.game.IsLastRound(),
		History:     b.history,
	}
	response, _ := json.Marshal(msg)
	return response, err
}

func (b *AcquireBridge) tilesData(pl acquireInterfaces.Player) map[string]bool {
	hnd := map[string]bool{}
	var coords string

	for _, tl := range pl.Tiles() {
		coords = strconv.Itoa(tl.Number()) + tl.Letter()
		hnd[coords] = b.game.IsTilePlayable(tl)
	}
	return hnd
}

func (b *AcquireBridge) corpsData() [7]corpData {
	var data [7]corpData
	for i, corp := range b.corporations {
		data[i] = corpData{
			Name:            corp.(*corporation.Corporation).Name(),
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
				Name:        p.(*player.Player).Name(),
				Active:      p.Active(),
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(i),
				InTurn:      b.isCurrentPlayer(i),
			})
		} else {
			ply = playerData{
				Name:        p.(*player.Player).Name(),
				Active:      p.Active(),
				Cash:        p.Cash(),
				OwnedShares: b.playersShares(n),
				InTurn:      b.isCurrentPlayer(n),
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
func (b *AcquireBridge) AddPlayer(name string) error {
	if len(b.players) == maximumPlayers {
		return errors.New(GameFull)
	}
	if b.GameStarted() {
		return errors.New(GameAlreadyStarted)
	}
	b.players = append(b.players, player.New(name))
	return nil
}

// RemovePlayer removes a player from the game
func (b *AcquireBridge) RemovePlayer(number int) error {
	if number < 0 || number > len(b.players) {
		return errors.New(InexistentPlayer)
	}
	b.players = append(b.players[:number], b.players[number+1:]...)
	return nil
}

// DeactivatePlayer removes a player from the game
func (b *AcquireBridge) DeactivatePlayer(number int) error {
	if number < 0 || number > len(b.players) {
		return errors.New(InexistentPlayer)
	}
	b.game.DeactivatePlayer(b.players[number])
	return nil
}

// StartGame starts a new Acquire game
func (b *AcquireBridge) StartGame() error {
	var err error
	if b.GameStarted() {
		err = errors.New(GameAlreadyStarted)
	}

	b.game, err = acquire.New(b.players, acquire.Optional{Corporations: b.corporations})
	if err == nil {
		b.history = append(b.history, fmt.Sprintf("%s is the starting player", b.currentPlayerName()))
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
	if name, ok := params.(string); ok {
		if bot, err := bots.Create(name); err == nil {
			return NewBotClient(bot, room), nil
		} else {
			return nil, err
		}
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
