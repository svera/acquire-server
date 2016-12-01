package acquirebridge

import (
	"errors"
	"strconv"

	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/tile"
)

func (b *AcquireBridge) playTile(clientName string, params playTileMessageParams) error {
	var err error
	var tl acquireInterfaces.Tile

	if tl, err = coordsToTile(params.Tile); err == nil {
		if err = b.game.PlayTile(tl); err == nil {
			b.history = append(b.history, i18n{
				Key: "game.history.played_tile",
				Arguments: map[string]string{
					"player": clientName,
					"tile":   params.Tile,
				},
			})
			return nil
		}
	}

	return err
}

func coordsToTile(tl string) (acquireInterfaces.Tile, error) {
	if len(tl) < 2 {
		return &tile.Tile{}, errors.New("Not a valid tile")
	}
	number, _ := strconv.Atoi(tl[:len(tl)-1])
	letter := string(tl[len(tl)-1 : len(tl)])
	return tile.New(number, letter), nil
}
