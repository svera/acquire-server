package room

import (
	"encoding/json"
	"errors"

	"github.com/svera/sackson-server/internal/events"
	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
)

func (r *Room) kickPlayerAction(m *interfaces.IncomingMessage) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed messages.KickPlayer
	if err = json.Unmarshal(m.Content, &parsed); err == nil {
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
	r.observer.Trigger(events.ClientOut{Client: cl, Reason: messages.ReasonPlayerKicked, Room: r})

	return nil
}
