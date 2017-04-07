package room

import (
	"encoding/json"
	"errors"
	"fmt"

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
			r.observer.Trigger("messageCreated", []interfaces.Client{m.Author}, response)
		}
	}
	return err
}

func (r *Room) addBot(level string) error {
	var err error
	var c interfaces.Client

	if c, err = r.gameBridge.AddBot(level, r); err == nil {
		c.SetName(fmt.Sprintf("Bot %d", r.clientCounter))
		if _, err = r.addClient(c); err == nil {
			go c.WritePump()
			go c.ReadPump(r.messages, r.unregister)
		}
	}
	return err
}
