package hub

import (
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) registerEvents() {
	h.observer.On(events.GameStarted{}, func(ev interface{}) {
		if event, ok := ev.(events.GameStarted); ok {
			message := h.createUpdatedRoomListMessage()

			gameClients := h.clients[event.Room.GameDriverName()]
			wg.Add(len(gameClients))
			for _, cl := range gameClients {
				go h.sendMessage(cl, message, messages.TypeRoomsList)
			}
		}
	})

	h.observer.On(events.GameStarted{}, func(ev interface{}) {
		if event, ok := ev.(events.GameStarted); ok {
			message := messages.GameStarted{
				PlayerTimeOut:  event.Room.PlayerTimeOut(),
				GameParameters: event.GameParameters,
			}

			wg.Add(len(event.Room.Clients()))
			for _, cl := range event.Room.Clients() {
				go h.sendMessage(cl, message, messages.TypeGameStarted)
			}

		}
	})

	h.observer.On(events.GameStatusUpdated{}, func(ev interface{}) {
		if event, ok := ev.(events.GameStatusUpdated); ok {
			wg.Add(1)
			go h.sendMessage(event.Client, event.Message, messages.TypeUpdateGameStatus, event.SequenceNumber)
		}
	})

	h.observer.On(events.RoomCreated{}, func(ev interface{}) {
		if event, ok := ev.(events.RoomCreated); ok {
			gameClients := h.clients[event.Room.GameDriverName()]
			wg.Add(len(gameClients))
			for _, cl := range gameClients {
				go h.sendMessage(cl, h.createUpdatedRoomListMessage(), messages.TypeRoomsList)
			}
		}
	})

	h.observer.On(events.RoomDestroyed{}, func(ev interface{}) {
		if event, ok := ev.(events.RoomDestroyed); ok {
			gameClients := h.clients[event.GameName]
			wg.Add(len(gameClients))
			for _, cl := range gameClients {
				go h.sendMessage(cl, h.createUpdatedRoomListMessage(), messages.TypeRoomsList)
			}
		}
	})

	h.observer.On(events.ClientRegistered{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientRegistered); ok {
			wg.Add(1)
			go h.sendMessage(event.Client, h.createUpdatedRoomListMessage(), messages.TypeRoomsList)
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
				h.destroyRoom(room.ID(), messages.ReasonRoomDestroyedNoClients)
			}
		}
	})

	h.observer.On(events.ClientOut{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientOut); ok {
			message := messages.ClientOut{
				Reason: event.Reason,
			}

			if len(event.Room.HumanClients()) == 0 && !event.Room.IsToBeDestroyed() {
				h.destroyRoom(event.Room.ID(), messages.ReasonRoomDestroyedNoClients)
			}
			wg.Add(1)
			go h.sendMessage(event.Client, message, messages.TypeClientOut)
		}
	})

	h.observer.On(events.ClientJoined{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientJoined); ok {
			message := messages.JoinedRoom{
				ClientNumber: event.ClientNumber,
				ID:           event.Client.Room().ID(),
				Owner:        event.Owner,
			}

			wg.Add(1)
			go h.sendMessage(event.Client, message, messages.TypeJoinedRoom)
		}
	})

	h.observer.On(events.ClientsUpdated{}, func(ev interface{}) {
		if event, ok := ev.(events.ClientsUpdated); ok {
			message := messages.CurrentPlayers{
				Values: event.PlayersData,
			}

			wg.Add(len(event.Clients))
			for _, cl := range event.Clients {
				go h.sendMessage(cl, message, messages.TypeCurrentPlayers)
			}
		}
	})

	h.observer.On(events.Error{}, func(ev interface{}) {
		if event, ok := ev.(events.Error); ok {
			message := messages.Error{
				Description: event.ErrorText,
			}

			wg.Add(1)
			go h.sendMessage(event.Client, message, messages.TypeError)
		}
	})
}
