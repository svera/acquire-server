package hub

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (h *Hub) joinRoomAction(m *interfaces.MessageFromClient) {
	var parsed interfaces.MessageJoinRoomParams
	if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if room, ok := h.rooms[parsed.Room]; ok {
			room.AddHuman(m.Author)
		} else {
			response := messages.New(interfaces.TypeMessageError, InexistentRoom)
			go h.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
		}

	}
}
