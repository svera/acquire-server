package room

import (
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) clientQuits(cl interfaces.Client) error {
	r.RemoveClient(cl)
	r.observer.Trigger(events.ClientOut{Client: cl, Reason: messages.ReasonPlayerQuitted, Room: r})
	return nil
}
