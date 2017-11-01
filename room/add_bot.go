package room

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/svera/sackson-server/client"
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
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
	var cast interface{}
	var c interfaces.Client

	if cast, err = r.gameDriver.CreateAI(level); err == nil {
		if ai, ok := cast.(interfaces.AI); ok {
			c = client.NewBot(ai, r)
			c.SetName(fmt.Sprintf("Bot %d", r.clientCounter))
			if _, err = r.addClient(c); err == nil {
				go c.WritePump()
				go c.ReadPump(r.messages, r.unregister)
			}
		} else {
			err = fmt.Errorf(DoesNotImplementAI)
		}
	}
	return err
}
