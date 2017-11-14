package hub

import (
	"errors"
	"log"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) terminateRoomAction(m *interfaces.IncomingMessage) error {
	if m.Author.Room() == nil {
		return errors.New(NotInARoom)
	}
	if m.Author != m.Author.Room().Owner() {
		return errors.New(Forbidden)
	}
	h.destroyRoom(m.Author.Room().ID(), messages.ReasonRoomDestroyedTerminated)
	return nil
}

func (h *Hub) destroyRoom(roomID string, reasonCode string) {
	mutex.Lock()
	defer mutex.Unlock()
	if h.configuration.Debug {
		log.Printf("Destroying room %s...", roomID)
	}
	if r, ok := h.rooms[roomID]; ok {
		r.ToBeDestroyed(true)
		if r.Timer() != nil {
			r.Timer().Stop()
		}
		h.expelClientsFromRoom(r, reasonCode)
		gameName := h.rooms[roomID].GameDriverName()
		delete(h.rooms, roomID)
		h.observer.Trigger(events.RoomDestroyed{GameName: gameName})

		if h.configuration.Debug {
			log.Printf("Room %s destroyed\n", roomID)
		}
	} else {
		log.Printf("No existe %s", roomID)
	}
}

func (h *Hub) expelClientsFromRoom(r interfaces.Room, reasonCode string) {
	for _, cl := range r.Clients() {
		if cl.IsBot() {
			cl.Close()
			if h.configuration.Debug {
				log.Printf("Bot %s destroyed", cl.Name())
			}
		} else {
			r.RemoveClient(cl)
			h.observer.Trigger(events.ClientOut{Client: cl, Reason: reasonCode, Room: r})
			if h.configuration.Debug {
				log.Printf("Client expelled from room %s\n", r.ID())
			}
		}
	}
}
