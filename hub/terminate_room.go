package hub

import (
	"log"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) terminateRoomAction(m *interfaces.IncomingMessage) {
	if m.Author != m.Author.Room().Owner() {
		return
	}
	h.destroyRoom(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedTerminated)
}

func (h *Hub) destroyRoom(roomID string, reasonCode string) {
	mutex.RLock()
	defer mutex.RUnlock()
	if r, ok := h.rooms[roomID]; ok {
		r.Timer().Stop()
		h.expelClientsFromRoom(r, reasonCode)
		delete(h.rooms, roomID)
		go h.emitter.Emit("messageCreated", h.clients, h.createUpdatedRoomListMessage())

		if h.configuration.Debug {
			log.Printf("Room %s destroyed\n", roomID)
		}
	}
}

func (h *Hub) expelClientsFromRoom(r interfaces.Room, reasonCode string) {
	response := messages.New(interfaces.TypeMessageClientOut, reasonCode)

	for _, cl := range r.Clients() {
		if cl != nil && cl.IsBot() {
			if h.configuration.Debug {
				log.Printf("Bot %s destroyed", cl.Name())
			}
			cl.Close()
		} else if cl != nil {
			go h.emitter.Emit("messageCreated", r.HumanClients(), response)
			cl.SetRoom(nil)
			cl.StopTimer()
		}
	}
}
