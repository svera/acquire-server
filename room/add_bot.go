package room

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/svera/sackson-server/client"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
)

func (r *Room) addBotAction(m *interfaces.IncomingMessage) error {
	var err error
	if m.Author != r.owner {
		return errors.New(Forbidden)
	}
	var parsed interfaces.MessageAddBotParams
	if err = json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if err = r.addBot(parsed.BotLevel); err != nil {
			response := messages.New(interfaces.TypeMessageError, err.Error())
			r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response, interfaces.TypeMessageError)
		}
	}
	return err
}

func (r *Room) addBot(level string) error {
	var err error
	var cast interface{}
	var c interfaces.Client

	if cast, err = r.gameBridge.CreateAI(level); err == nil {
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
