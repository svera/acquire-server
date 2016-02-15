package acquirebridge

import (
	"github.com/svera/acquire"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/fsm"
	"github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
)

func (b *acquireBridge) NewGameMergeTest() {
	bd := board.New()
	ts := tileset.NewStub()
	b.corporations = createCorporations()
	tiles := []interfaces.Tile{
		tile.New(5, "E"),
		tile.New(6, "E"),
	}
	tiles2 := []interfaces.Tile{
		tile.New(8, "E"),
		tile.New(9, "E"),
		tile.New(10, "E"),
	}

	ts.DiscardTile(tiles[0])
	ts.DiscardTile(tiles[1])
	ts.DiscardTile(tiles2[0])
	ts.DiscardTile(tiles2[1])
	ts.DiscardTile(tiles2[2])
	bd.SetOwner(b.corporations[0], tiles)
	bd.SetOwner(b.corporations[1], tiles2)
	b.corporations[0].Grow(2)
	b.corporations[1].Grow(3)

	b.game, _ = acquire.New(
		bd,
		b.players,
		b.corporations,
		ts,
		&fsm.PlayTile{},
	)

	b.players[0].DiscardTile(b.players[0].Tiles()[0])
	b.players[0].PickTile(tile.New(7, "E"))
	b.players[0].AddShares(b.corporations[0], 5)
	b.players[1].AddShares(b.corporations[0], 5)
}

func (b *acquireBridge) NewGameTiedMergeTest() {
	bd := board.New()
	ts := tileset.NewStub()
	b.corporations = createCorporations()
	tiles := []interfaces.Tile{
		tile.New(4, "E"),
		tile.New(5, "E"),
		tile.New(6, "E"),
	}
	tiles2 := []interfaces.Tile{
		tile.New(8, "E"),
		tile.New(9, "E"),
		tile.New(10, "E"),
	}

	ts.DiscardTile(tiles[0])
	ts.DiscardTile(tiles[1])
	ts.DiscardTile(tiles[2])
	ts.DiscardTile(tiles2[0])
	ts.DiscardTile(tiles2[1])
	ts.DiscardTile(tiles2[2])
	bd.SetOwner(b.corporations[0], tiles)
	bd.SetOwner(b.corporations[1], tiles2)
	b.corporations[0].Grow(3)
	b.corporations[1].Grow(3)

	b.game, _ = acquire.New(
		bd,
		b.players,
		b.corporations,
		ts,
		&fsm.PlayTile{},
	)

	b.players[0].DiscardTile(b.players[0].Tiles()[0])
	b.players[0].PickTile(tile.New(7, "E"))
	b.players[0].AddShares(b.corporations[0], 5)
	b.players[1].AddShares(b.corporations[0], 5)
	b.players[0].AddShares(b.corporations[1], 3)
	b.players[1].AddShares(b.corporations[1], 3)
}
