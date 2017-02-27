// Package player contains the model Player and attahced methods which manages corporations in game
package player

import acquirePlayer "github.com/svera/acquire/player"

// Player holds data related to players
type Player struct {
	*acquirePlayer.Player
	name string
}

// New initialises and returns a new instance of Player
func New(name string, number int) *Player {
	return &Player{
		acquirePlayer.New(number),
		name,
	}
}

// Name returns the player name
func (p *Player) Name() string {
	return p.name
}
