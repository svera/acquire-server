package acquirebridge

import (
	"encoding/json"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/tbg-server/client"
	serverInterfaces "github.com/svera/tbg-server/interfaces"
	"log"
	"strconv"
)

const (
	maxMessageSize = 1024 * 1024
)

// Bot is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type BotClient struct {
	name     string
	incoming chan []byte // Channel storing incoming messages
	botTurn  chan statusMessage
	owner    bool
	bot      bot
}

// NewBot returns a new Bot instance
func NewBotClient(b bot) serverInterfaces.Client {
	return &BotClient{
		incoming: make(chan []byte, maxMessageSize),
		botTurn:  make(chan statusMessage),
		bot:      b,
	}
}

// ReadPump reads input from the user and writes it to the passed channel,
// with usually belongs to the hub
func (c *BotClient) ReadPump(cnl interface{}, unregister chan serverInterfaces.Client) {
	var msg *client.Message
	channel := cnl.(chan *client.Message)
	defer func() {
		unregister <- c
	}()

	for {
		select {
		case parsed := <-c.botTurn:
			c.updateBot(parsed)
			switch parsed.State {
			case acquireInterfaces.PlayTileStateName:
				msg = c.playTile()
			}
		}
		channel <- msg
	}

}

func (c *BotClient) updateBot(parsed statusMessage) {
	var hand []acquireInterfaces.Tile
	for _, handData := range parsed.PlayerInfo.Hand {
		tl, _ := coordsToTile(handData.Coords)
		hand = append(hand, tl)
	}
	c.bot.SetTiles(hand)
}

func (c *BotClient) playTile() *client.Message {
	tl := c.bot.PlayTile()
	log.Println(strconv.Itoa(tl.Number()) + tl.Letter())
	params := playTileMessageParams{
		Tile: strconv.Itoa(tl.Number()) + tl.Letter(),
	}
	ser, _ := json.Marshal(params)
	return &client.Message{
		Author: c,
		Content: client.MessageContent{
			Type:   messageTypePlayTile,
			Params: ser,
		},
	}
}

// WritePump sends data to the user
func (c *BotClient) WritePump() {
	for {
		select {
		case message, ok := <-c.incoming:
			if !ok {
				return
			}
			var parsed statusMessage
			if err := json.Unmarshal(message, &parsed); err == nil {
				if parsed.PlayerInfo.Enabled {
					c.botTurn <- parsed
				}
			} else {
				log.Println(err)
			}
		}
	}
}

func (c *BotClient) Incoming() chan []byte {
	return c.incoming
}

// Owner always return false for bots
func (c *BotClient) Owner() bool {
	return false
}

// SetOwner doesn't change Owner status in a bot, as bots cannot be owners
func (c *BotClient) SetOwner(v bool) serverInterfaces.Client {
	return c
}

func (c *BotClient) Name() string {
	return c.name
}

func (c *BotClient) SetName(v string) serverInterfaces.Client {
	c.name = v
	return c
}
