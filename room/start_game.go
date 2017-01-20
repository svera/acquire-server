package room

import (
	"encoding/json"
	"errors"

	"github.com/svera/sackson-server/interfaces"
)

func (r *Room) startGameAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageStartGameParams

	if m.Author != r.owner {
		return errors.New(Forbidden)
	}

	if err := json.Unmarshal(m.Content.Params, &parsed); err != nil {
		return err
	}

	if err := r.startGame(); err != nil {
		return err
	}

	for n, cl := range r.clients {
		st, _ := r.gameBridge.Status(n)
		go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, st)
	}

	r.playerTimeOut = parsed.PlayerTimeout
	r.changePlayerSetTimer()

	go r.emitter.Emit(GameStarted)
	return nil
}

func (r *Room) startGame() error {
	return r.gameBridge.StartGame()
}
