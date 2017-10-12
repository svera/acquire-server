package room

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
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

	if err = r.gameDriver.StartGame(r.mapPlayerNames()); err != nil {
		return err
	}

	if err = r.sendInitialMessage(); err != nil {
		return err
	}

	r.changeClientsInTurn()

	r.observer.Trigger(events.GameStarted{Room: r, GameParameters: m.Content.Params})
	return err
}

func (r *Room) sendInitialMessage() error {
	var status interface{}
	var err error

	r.updateSequenceNumber++
	for n, cl := range r.clients {
		if status, err = r.gameDriver.Status(n); err != nil {
			return err
		}
		r.setUpTimeOut(cl)
		r.observer.Trigger(events.GameStatusUpdated{Client: cl, Message: status, SequenceNumber: r.updateSequenceNumber})
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
	r.RemoveClient(cl)
	r.observer.Trigger(events.ClientOut{Client: cl, Reason: interfaces.ReasonPlayerTimedOut, Room: r})
}

func (r *Room) mapPlayerNames() map[int]string {
	names := map[int]string{}
	for n, cl := range r.clients {
		names[n] = cl.Name()
	}
	return names
}
