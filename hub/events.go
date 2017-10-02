package hub

import (
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) registerEvents() {
	h.observer.On(events.GameStarted, func(args ...interface{}) {
		message := h.createUpdatedRoomListMessage()

		wg.Add(len(h.clients))
		for _, cl := range h.clients {
			go h.sendMessage(cl, message, interfaces.TypeMessageRoomsList)
		}
	})

	h.observer.On(events.GameStatusUpdated, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		message := args[1]
		sequenceNumber := args[2].(int)

		wg.Add(1)
		go h.sendMessage(client, message, interfaces.TypeMessageUpdateGameStatus, sequenceNumber)
	})

	h.observer.On(events.RoomCreated, func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)

		wg.Add(len(clients))
		for _, cl := range clients {
			go h.sendMessage(cl, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
		}
	})

	h.observer.On(events.RoomDestroyed, func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)

		wg.Add(len(clients))
		for _, cl := range clients {
			go h.sendMessage(cl, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
		}
	})

	h.observer.On(events.ClientRegistered, func(args ...interface{}) {
		client := args[0].(interfaces.Client)

		wg.Add(1)
		go h.sendMessage(client, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)
	})

	h.observer.On(events.ClientUnregistered, func(args ...interface{}) {
		if len(args) > 0 {
			room := args[0].(interfaces.Room)
			if len(room.HumanClients()) == 0 {
				h.destroyRoom(room.ID(), interfaces.ReasonRoomDestroyedNoClients)
			}
		}
	})

	h.observer.On(events.ClientOut, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		reasonCode := args[1].(string)
		room := args[2].(interfaces.Room)
		message := messages.New(interfaces.TypeMessageClientOut, reasonCode)

		if len(room.HumanClients()) == 0 && reasonCode != interfaces.ReasonRoomDestroyedTerminated {
			h.destroyRoom(room.ID(), interfaces.ReasonRoomDestroyedNoClients)
		}
		wg.Add(1)
		go h.sendMessage(client, message, interfaces.TypeMessageClientOut)
	})

	h.observer.On(events.ClientJoined, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		clientNumber := args[1].(int)
		owner := args[2].(bool)

		message := messages.New(interfaces.TypeMessageJoinedRoom, clientNumber, client.Room().ID(), owner)

		wg.Add(1)
		go h.sendMessage(client, message, interfaces.TypeMessageJoinedRoom)
	})

	h.observer.On(events.ClientsUpdated, func(args ...interface{}) {
		clients := args[0].([]interfaces.Client)
		playersData := args[1].(map[string]interfaces.PlayerData)

		message := messages.New(interfaces.TypeMessageCurrentPlayers, playersData)

		wg.Add(len(clients))
		for _, cl := range clients {
			go h.sendMessage(cl, message, interfaces.TypeMessageCurrentPlayers)
		}
	})

	h.observer.On(events.Error, func(args ...interface{}) {
		client := args[0].(interfaces.Client)
		errorText := args[1]
		errorMessage := messages.New(interfaces.TypeMessageError, errorText)

		wg.Add(1)
		go h.sendMessage(client, errorMessage, interfaces.TypeMessageError)
	})
}
