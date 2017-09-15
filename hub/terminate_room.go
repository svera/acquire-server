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
			h.observer.Trigger(events.ClientOut, cl, reasonCode)
			if h.configuration.Debug {
				log.Printf("Client expeled from room %s\n", cl.Room().ID())
			}
			cl.SetRoom(nil)
			cl.StopTimer()
		}
	}
}
