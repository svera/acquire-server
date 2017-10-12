package hub

import (
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) registerEvents() {
	h.observer.On(events.GameStarted{}, func(ev interface{}) {
		if _, ok := ev.(events.GameStarted); ok {
			message := h.createUpdatedRoomListMessage()

			wg.Add(len(h.clients))
			for _, cl := range h.clients {
				go h.sendMessage(cl, message, interfaces.TypeMessageRoomsList)
			}
		}
	})

	h.observer.On(events.GameStarted{}, func(ev interface{}) {
		if event, ok := ev.(events.GameStarted); ok {
			message := messages.New(interfaces.TypeMessageGameStarted, event.GameParameters)

			wg.Add(len(event.Room.Clients()))
			for _, cl := range event.Room.Clients() {
				go h.sendMessage(cl, message, interfaces.TypeMessageGameStarted)
			}

		}
	})

	h.observer.On(events.GameStatusUpdated{}, func(ev interface{}) {
		if event, ok := ev.(events.GameStatusUpdated); ok {
			wg.Add(1)
			go h.sendMessage(event.Client, event.Message, interfaces.TypeMessageUpdateGameStatus, event.SequenceNumber)
		}
	})

	h.observer.On(events.RoomCreated{}, func(ev interface{}) {
		if _, ok := ev.(events.RoomCreated); ok {
			wg.Add(len(h.clients))
			for _, cl := range h.clients {
				go h.sendMessage(cl, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
			}
		}
	})

	h.observer.On(events.RoomDestroyed{}, func(ev interface{}) {
		if _, ok := ev.(events.RoomDestroyed); ok {
			wg.Add(len(h.clients))
			for _, cl := range h.clients {
				go h.sendMessage(cl, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
			}
		}
	})

	h.observer.On(events.ClientRegistered{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientRegistered); ok {
			wg.Add(1)
			go h.sendMessage(event.Client, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
		}
	})

	h.observer.On(events.ClientUnregistered{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientUnregistered); ok {
			var room interfaces.Room
			if room = event.Client.Room(); room == nil {
				return
			}

			room.RemoveClient(event.Client)
			if len(room.HumanClients()) == 0 && !room.IsToBeDestroyed() {
				h.destroyRoom(room.ID(), interfaces.ReasonRoomDestroyedNoClients)
			}
		}
	})

	h.observer.On(events.ClientOut{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientOut); ok {
			message := messages.New(interfaces.TypeMessageClientOut, event.Reason)

			if len(event.Room.HumanClients()) == 0 && !event.Room.IsToBeDestroyed() {
				h.destroyRoom(event.Room.ID(), interfaces.ReasonRoomDestroyedNoClients)
			}
			wg.Add(1)
			go h.sendMessage(event.Client, message, interfaces.TypeMessageClientOut)
		}
	})

	h.observer.On(events.ClientJoined{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientJoined); ok {
			message := messages.New(interfaces.TypeMessageJoinedRoom, event.ClientNumber, event.Client.Room().ID(), event.Owner)

			wg.Add(1)
			go h.sendMessage(event.Client, message, interfaces.TypeMessageJoinedRoom)
		}
	})

	h.observer.On(events.ClientsUpdated{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientsUpdated); ok {
			message := messages.New(interfaces.TypeMessageCurrentPlayers, event.PlayersData)

			wg.Add(len(event.Clients))
			for _, cl := range event.Clients {
				go h.sendMessage(cl, message, interfaces.TypeMessageCurrentPlayers)
			}
		}
	})

	h.observer.On(events.Error{}, func(ev interface{}) {
		if event, ok := ev.(events.Error); ok {
			errorMessage := messages.New(interfaces.TypeMessageError, event.ErrorText)

			wg.Add(1)
			go h.sendMessage(event.Client, errorMessage, interfaces.TypeMessageError)
		}
	})
}
