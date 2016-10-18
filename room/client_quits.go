package room

import "github.com/svera/sackson-server/interfaces"

func (r *Room) clientQuits(cl interfaces.Client) error {
	r.RemoveClient(cl)
	response := newMessage(interfaces.TypeMessageRoomDestroyed, "qui")
	go r.emitter.Emit("messageCreated", []interfaces.Client{cl}, response)
	return nil
}
