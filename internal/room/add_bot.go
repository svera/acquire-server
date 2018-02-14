package room

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/svera/sackson-server/api"
	"github.com/svera/sackson-server/internal/client"
	"github.com/svera/sackson-server/internal/events"
	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
)

func (r *Room) addBotAction(m *interfaces.IncomingMessage) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed messages.AddBot
	if err = json.Unmarshal(m.Content, &parsed); err == nil {
		if err = r.addBot(parsed.BotLevel); err != nil {
			r.observer.Trigger(events.Error{Client: m.Author, ErrorText: err.Error()})
		}
	}
	return err
}

func (r *Room) addBot(level string) error {
	var err error
	var ai api.AI
	var c interfaces.Client

	if ai, err = r.gameDriver.CreateAI(level); err == nil {
		c = client.NewBot(ai, r, r.observer)
		c.SetName(fmt.Sprintf("Bot %d", r.clientCounter))
		if _, err = r.addClient(c); err == nil {
			go c.WritePump()
			go c.ReadPump(r.messages, r.unregister)
		}
	}
	return err
}
