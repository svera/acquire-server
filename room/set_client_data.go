package room

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) setClientDataAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageSetClientDataParams
	var err error

	if err = json.Unmarshal(m.Content.Params, &parsed); err != nil {
		return err
	}
	m.Author.SetName(parsed.Name)
	response := messages.New(interfaces.TypeMessageCurrentPlayers, r.playersData())
	r.observer.Trigger("messageCreated", mapToSlice(r.clients), response, interfaces.TypeMessageCurrentPlayers)

	return nil
}
