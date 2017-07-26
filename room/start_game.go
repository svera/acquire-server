package room

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) startGameAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageStartGameParams
	var err error

	if m.Author != r.owner {
		return errors.New(Forbidden)
	}

	if err = json.Unmarshal(m.Content.Params, &parsed); err != nil {
		return err
	}
	r.playerTimeOut = parsed.PlayerTimeout

	if err = r.gameBridge.StartGame(r.clients); err != nil {
		return err
	}

	if err = r.sendInitialMessage(); err != nil {
		return err
	}

	r.changeClientsInTurn()

	r.observer.Trigger(GameStarted)
	return err
}

func (r *Room) sendInitialMessage() error {
	var status interface{}
	var err error

	r.updateSequenceNumber++
	for n, cl := range r.clients {
		if status, err = r.gameBridge.Status(n); err != nil {
			return err
		}
		r.setUpTimeOut(cl)
		r.observer.Trigger(GameStatusUpdated, cl, status, r.updateSequenceNumber)
	}
	return nil
}

// Sets up a timer that will execute when the defined player timeout is reached.
func (r *Room) setUpTimeOut(cl interfaces.Client) {
	if r.playerTimeOut > 0 && !cl.IsBot() {
		cl.SetTimer(time.AfterFunc(time.Second*r.playerTimeOut, func() {
			if r.configuration.Debug {
				log.Printf("Client '%s' timed out", cl.Name())
			}
			r.timeoutPlayer(cl)
		}))
	}
}

func (r *Room) timeoutPlayer(cl interfaces.Client) {
	response := messages.New(interfaces.TypeMessageClientOut, interfaces.ReasonPlayerTimedOut)

	r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response, interfaces.TypeMessageClientOut)
	r.RemoveClient(cl)
}
