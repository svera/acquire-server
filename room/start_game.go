package room

import (
	"encoding/json"
	"errors"
	"log"
	"time"

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

	r.playerTimeOut = parsed.PlayerTimeout

	for n, cl := range r.clients {
		st, _ := r.gameBridge.Status(n)
		r.setUpTimeOut(cl)
		go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, st)
	}

	r.changePlayerSetTimer()

	go r.emitter.Emit(GameStarted)
	return nil
}

func (r *Room) startGame() error {
	return r.gameBridge.StartGame()
}

// Sets up a timer that will execute when the defined player timeout is reached.
func (r *Room) setUpTimeOut(cl interfaces.Client) {
	if r.playerTimeOut > 0 {
		cl.SetTimer(time.AfterFunc(time.Second*r.playerTimeOut, func() {
			if r.configuration.Debug {
				log.Printf("Client '%s' timed out", cl.Name())
			}
			r.timeoutPlayer(cl)
		}))
	}
}
