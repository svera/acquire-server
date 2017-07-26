package room

import (
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) clientQuits(cl interfaces.Client) error {
	r.RemoveClient(cl)
	response := messages.New(interfaces.TypeMessageClientOut, interfaces.ReasonPlayerQuitted)
	r.observer.Trigger("messageCreated", []interfaces.Client{cl}, response, interfaces.TypeMessageClientOut)
	return nil
}
