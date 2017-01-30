package hub

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/svera/sackson-server/interfaces"
)

func (h *Hub) joinRoomAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageJoinRoomParams
	var err error
	var room interfaces.Room
	var ok bool

	if err = json.Unmarshal(m.Content.Params, &parsed); err != nil {
		return err
	}
	if room, ok = h.rooms[parsed.Room]; !ok {
		return errors.New(InexistentRoom)
	}

	if strings.TrimSpace(parsed.ClientName) != "" {
		m.Author.SetName(parsed.ClientName)
	}
	room.AddHuman(m.Author)
	return nil
}
