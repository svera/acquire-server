package room

import (
	"encoding/json"

	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) setClientDataAction(m *interfaces.IncomingMessage) error {
	var parsed messages.SetClientDataParams
	var err error

	if err = json.Unmarshal(m.Content, &parsed); err != nil {
		return err
	}
	m.Author.SetName(parsed.Name)
	r.observer.Trigger(events.ClientsUpdated{Clients: mapToSlice(r.clients), PlayersData: r.playersData()})

	return nil
}
