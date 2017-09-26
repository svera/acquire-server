package room

import (
	"encoding/json"
	"errors"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
)

func (r *Room) kickPlayerAction(m *interfaces.IncomingMessage) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed interfaces.MessageKickPlayerParams
	if err = json.Unmarshal(m.Content.Params, &parsed); err == nil {
		return r.kickClient(parsed.PlayerNumber)
	}
	return err
}

func (r *Room) kickClient(number int) error {
	if _, exist := r.clients[number]; !exist {
		return errors.New(InexistentClient)
	}
	cl := r.clients[number]
	if cl == r.owner {
		return errors.New(OwnerNotRemovable)
	}
	cl.SetRoom(nil)
	r.RemoveClient(r.clients[number])
	r.observer.Trigger(events.ClientOut, cl, interfaces.ReasonPlayerKicked, r)

	return nil
}
