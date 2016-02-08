package bridge

import (
	"github.com/svera/acquire"
	"github.com/svera/acquire/board"
	"github.com/svera/acquire/fsm"
	"github.com/svera/acquire/interfaces"
	"github.com/svera/acquire/tile"
	"github.com/svera/acquire/tileset"
)

func (b *Bridge) NewGameMergeTest(players []interfaces.Player) {
	bd := board.New()
	ts := tileset.NewStub()
	corps := createCorporations()
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
	bd.SetOwner(corps[0], tiles)
	bd.SetOwner(corps[1], tiles2)
	corps[0].Grow(2)
	corps[1].Grow(3)

	b.game, _ = acquire.New(
		bd,
		players,
		corps,
		tileset.New(),
		&fsm.PlayTile{},
	)

	players[0].DiscardTile(players[0].Tiles()[0])
	players[0].PickTile(tile.New(7, "E"))
	players[0].AddShares(corps[0], 5)
	players[1].AddShares(corps[0], 5)
}

func (b *Bridge) NewGameTiedMergeTest(players []interfaces.Player) {
	bd := board.New()
	ts := tileset.NewStub()
	corps := createCorporations()
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
	bd.SetOwner(corps[0], tiles)
	bd.SetOwner(corps[1], tiles2)
	corps[0].Grow(3)
	corps[1].Grow(3)

	b.game, _ = acquire.New(
		bd,
		players,
		corps,
		tileset.New(),
		&fsm.PlayTile{},
	)

	players[0].DiscardTile(players[0].Tiles()[0])
	players[0].PickTile(tile.New(7, "E"))
	players[0].AddShares(corps[0], 5)
	players[1].AddShares(corps[0], 5)
	players[0].AddShares(corps[1], 3)
	players[1].AddShares(corps[1], 3)
}