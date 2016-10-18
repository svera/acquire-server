package hub

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

func (h *Hub) joinRoomAction(m *interfaces.MessageFromClient) {
	var parsed interfaces.MessageJoinRoomParams
	if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if room, ok := h.rooms[parsed.Room]; ok {
			room.AddHuman(m.Author)
		} else {
			res := &interfaces.MessageError{
				Type:    "err",
				Content: InexistentRoom,
			}
			response, _ := json.Marshal(res)
			go h.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
		}

	}
}
