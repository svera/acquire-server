package hub

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
)

func (h *Hub) joinRoomAction(m *interfaces.IncomingMessage) error {
	var parsed messages.JoinRoom
	var err error
	var room interfaces.Room
	var ok bool

	if err = json.Unmarshal(m.Content, &parsed); err != nil {
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
