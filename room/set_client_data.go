package room

import (
	"encoding/json"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
)

func (r *Room) setClientDataAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageSetClientDataParams
	var err error

	if err = json.Unmarshal(m.Content.Params, &parsed); err != nil {
		return err
	}
	m.Author.SetName(parsed.Name)
	r.observer.Trigger(events.ClientsUpdated{Clients: mapToSlice(r.clients), PlayersData: r.playersData()})

	return nil
}
