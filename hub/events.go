package hub

import (
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/room"
)

// Not all messages need to be processed in a specific order, in that case
// we just use this noSequenceNumber to indicate this
const noSequenceNumber = 0

func (h *Hub) registerEvents() {
	h.observer.On(room.GameStarted, func(args ...interface{}) {
		message := h.createUpdatedRoomListMessage()

		wg.Add(len(h.clients))
		for _, cl := range h.clients {
			go h.sendMessage(cl, message, interfaces.TypeMessageRoomsList, noSequenceNumber)
		}
	})

	h.observer.On("messageCreated", func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)
		message := args[1]
		typeName := args[2].(string)

		wg.Add(len(clients))
		for _, cl := range clients {
			go h.sendMessage(cl, message, typeName, noSequenceNumber)
		}
	})

	h.observer.On(room.GameStatusUpdated, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		message := args[1]
		sequenceNumber := args[2].(int)

		wg.Add(1)
		go h.sendMessage(client, message, interfaces.TypeMessageUpdateGameStatus, sequenceNumber)
	})

	h.observer.On(room.ClientOut, func(args ...interface{}) {
		r := args[0].(interfaces.Room)
		if len(r.HumanClients()) == 0 {
			wg.Add(1)
			go h.destroyRoomConcurrently(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
	})
}
