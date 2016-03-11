package acquirebridge

import (
	"github.com/svera/acquire/interfaces"
)

type bot interface {
	SetCash(cash int)
	SetTiles(tiles []interfaces.Tile)
	PlayTile() interfaces.Tile
}
