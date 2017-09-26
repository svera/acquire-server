package hub

import (
	"errors"
	"log"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
)

func (h *Hub) terminateRoomAction(m *interfaces.IncomingMessage) error {
	if m.Author.Room() == nil {
		return errors.New(NotInARoom)
	}
	if m.Author != m.Author.Room().Owner() {
		return errors.New(Forbidden)
	}
	h.destroyRoom(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedTerminated)
	return nil
}

func (h *Hub) destroyRoom(roomID string, reasonCode string) {
	mutex.Lock()
	defer mutex.Unlock()

	if h.configuration.Debug {
		log.Printf("Destroying room %s...", roomID)
	}
	if r, ok := h.rooms[roomID]; ok {
		r.Timer().Stop()
		h.expelClientsFromRoom(r, reasonCode)
		delete(h.rooms, roomID)
		h.observer.Trigger(events.RoomDestroyed, h.clients)

		if h.configuration.Debug {
			log.Printf("Room %s destroyed\n", roomID)
		}
	}
}

func (h *Hub) expelClientsFromRoom(r interfaces.Room, reasonCode string) {
	for _, cl := range r.Clients() {
		if cl != nil && cl.IsBot() {
			if h.configuration.Debug {
				log.Printf("Bot %s destroyed", cl.Name())
			}
			cl.Close()
		} else if cl != nil {
			r.RemoveClient(cl)
			h.observer.Trigger(events.ClientOut, cl, reasonCode, r)
			if h.configuration.Debug {
				log.Printf("Client expelled from room %s\n", r.ID())
			}
		}
	}
}
