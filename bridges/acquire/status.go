package acquirebridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/sackson-server/bridges/acquire/corporation"
	"github.com/svera/sackson-server/bridges/acquire/player"
)

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

func (b *AcquireBridge) boardOwnership() map[string]string {
	cells := make(map[string]string)
	var letters = [9]string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for number := 1; number < 13; number++ {
		for _, letter := range letters {
			cell := b.game.Board().Cell(number, letter)
			if cell.Type() == "corporation" {
				cells[strconv.Itoa(number)+letter] = fmt.Sprintf("%d", cell.(*corporation.Corporation).Index())
			} else {
				cells[strconv.Itoa(number)+letter] = cell.Type()
			}
		}
	}

	return cells
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
			Tied:            false,
		}
	}

	for _, corp := range b.game.TiedCorps() {
		data[corp.(*corporation.Corporation).Index()].Tied = true
	}
	return data
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
