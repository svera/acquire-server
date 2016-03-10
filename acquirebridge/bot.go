package acquirebridge

import (
	"encoding/json"
	"github.com/svera/acquire-server/client"
	serverInterfaces "github.com/svera/acquire-server/interfaces"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"log"
)

const (
	maxMessageSize = 1024 * 1024
)

// Bot is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type Bot struct {
	name       string
	incoming   chan []byte // Channel storing incoming messages
	parsedChan chan statusMessage
	owner      bool
}

// NewBot returns a new Bot instance
func NewBot() serverInterfaces.Client {
	return &Bot{
		incoming:   make(chan []byte, maxMessageSize),
		parsedChan: make(chan statusMessage),
	}
}

// ReadPump reads input from the user and writes it to the passed channel,
// with usually belongs to the hub
func (c *Bot) ReadPump(cnl interface{}, unregister chan serverInterfaces.Client) {
	channel := cnl.(chan *client.Message)
	defer func() {
		unregister <- c
	}()

	for {
		select {
		case parsed := <-c.parsedChan:
			if parsed.State == acquireInterfaces.PlayTileStateName {
				params := playTileMessageParams{
					Tile: parsed.PlayerInfo.Hand[0].Coords,
				}
				ser, _ := json.Marshal(params)
				cnt := client.MessageContent{
					Type:   "ply",
					Params: ser,
				}
				msg := &client.Message{
					Author:  c,
					Content: cnt,
				}

				channel <- msg
			}
		}
	}

}

// WritePump sends data to the user
func (c *Bot) WritePump() {
	for {
		select {
		case message, ok := <-c.incoming:
			if !ok {
				return
			}
			var parsed statusMessage
			if err := json.Unmarshal(message, &parsed); err == nil {
				if parsed.PlayerInfo.Enabled {
					c.parsedChan <- parsed
				}
			} else {
				log.Println(err)
			}
		}
	}
}

func (c *Bot) Incoming() chan []byte {
	return c.incoming
}

// Owner always return false for bots
func (c *Bot) Owner() bool {
	return false
}

// SetOwner doesn't change Owner status in a bot, as bots cannot be owners
func (c *Bot) SetOwner(v bool) serverInterfaces.Client {
	return c
}

func (c *Bot) Name() string {
	return c.name
}

func (c *Bot) SetName(v string) serverInterfaces.Client {
	c.name = v
	return c
}
