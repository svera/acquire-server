package room

import (
	"github.com/svera/sackson-server/internal/events"
	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
)

func (r *Room) clientQuits(cl interfaces.Client) error {
	r.RemoveClient(cl)
	r.observer.Trigger(events.ClientOut{Client: cl, Reason: messages.ReasonPlayerQuitted, Room: r})
	return nil
}
