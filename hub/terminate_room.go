package hub

import (
	"encoding/json"
	"log"

	"github.com/svera/sackson-server/interfaces"
)

func (h *Hub) terminateRoomAction(m *interfaces.MessageFromClient) {
	if m.Author != m.Author.Room().Owner() {
		return
	}
	h.destroyRoom(m.Author.Room().ID(), interfaces.ReasonRoomDestroyedTerminated)
}

func (h *Hub) destroyRoom(roomID string, reasonCode string) {
	r := h.rooms[roomID]
	r.Timer().Stop()

	h.expelClientsFromRoom(r, reasonCode)

	mapLock.RLock()
	delete(h.rooms, roomID)
	mapLock.RUnlock()
	go h.emitter.Emit("messageCreated", h.clients, h.createUpdatedRoomListMessage())

	if h.debug {
		log.Printf("Room %s destroyed\n", roomID)
	}
}

func (h *Hub) expelClientsFromRoom(r interfaces.Room, reasonCode string) {
	msg := interfaces.MessageRoomDestroyed{
		Type:   interfaces.TypeMessageRoomDestroyed,
		Reason: reasonCode,
	}
	response, _ := json.Marshal(msg)

	for _, cl := range r.Clients() {
		if cl != nil && cl.IsBot() {
			if h.debug {
				log.Printf("Bot %s destroyed", cl.Name())
			}
			cl.Close()
		} else if cl != nil {
			go h.emitter.Emit("messageCreated", r.Clients(), response)
			cl.SetRoom(nil)
			cl.StopTimer()
		}
	}
}

func (h *Hub) createUpdatedRoomListMessage() []byte {
	msgRoomList := interfaces.MessageRoomsList{
		Type:   interfaces.TypeMessageRoomsList,
		Values: h.getWaitingRoomsIds(),
	}
	response, _ := json.Marshal(msgRoomList)
	return response
}
