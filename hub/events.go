package hub

import (
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/room"
)

func (h *Hub) registerEvents() {
	h.observer.On(room.GameStarted, func(args ...interface{}) {
		message := h.createUpdatedRoomListMessage()

		wg.Add(len(h.clients))
		for _, cl := range h.clients {
			go h.sendMessage(cl, message)
		}
	})

	h.observer.On("messageCreated", func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)
		message := args[1]

		wg.Add(len(clients))
		for _, cl := range clients {
			go h.sendMessage(cl, message)
		}
	})

	h.observer.On(room.GameStatusUpdated, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		message := args[1]

		wg.Add(1)
		go h.sendMessage(client, message)
	})

	h.observer.On(room.ClientOut, func(args ...interface{}) {
		r := args[0].(interfaces.Room)
		if len(r.HumanClients()) == 0 {
			wg.Add(1)
			go h.destroyRoomConcurrently(r.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
	})
}
