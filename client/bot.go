package client

import (
	"encoding/json"
	"log"
<<<<<<< Updated upstream
=======
	"sort"
>>>>>>> Stashed changes
	"time"

	"github.com/svera/sackson-server/interfaces"
)

// BotClient is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type BotClient struct {
	name         string
	incoming     chan []byte // Channel storing incoming messages
	endReadPump  chan struct{}
	endWritePump chan struct{}
	botTurn      chan struct{}
	ai           interfaces.AI
	room         interfaces.Room
}

// NewBot returns a new Bot instance
func NewBot(ai interfaces.AI, room interfaces.Room) interfaces.Client {
	return &BotClient{
		incoming:     make(chan []byte, maxMessageSize),
		endReadPump:  make(chan struct{}),
		endWritePump: make(chan struct{}),
		botTurn:      make(chan struct{}),
		ai:           ai,
		room:         room,
	}
}

// ReadPump listens to the botTurn channel (see the WritePump function) and, when
// an update message comes this way, updates the bot game status information
// and gets its next play, sending it back to the hub
func (c *BotClient) ReadPump(cnl interface{}, unregister chan interfaces.Client) {
	channel := cnl.(chan *interfaces.IncomingMessage)
	defer func() {
		unregister <- c
	}()

	for {
		select {
		case <-c.endReadPump:
			return

		case <-c.botTurn:
			log.Printf("BOT TURN REACHED\n")
			msgType, content := c.ai.Play()
			msg := &interfaces.IncomingMessage{
				Author: c,
				Content: interfaces.IncomingMessageContent{
					Type:   msgType,
					Params: content,
				},
			}
			channel <- msg
		}
	}

}

// WritePump gets updates from the hub
func (c *BotClient) WritePump() {
	var parsed interfaces.OutgoingMessage
	var err error

	for {
		select {
		case <-c.endWritePump:
			return

		case message, ok := <-c.incoming:
			if !ok {
				return
			}

			if err = json.Unmarshal(message, &parsed); err == nil {
				if err = c.ai.FeedGameStatus(parsed.Content); err == nil {
					if c.ai.IsInTurn() {
						c.botTurn <- struct{}{}
					}
				}
			}
		}
	}
}

func (c *BotClient) feedPendingUpdatesInOrder() {
	for _, seq := range c.getSortedUpdatesBufferKeys() {
		c.ai.FeedGameStatus(c.updatesBuffer[seq])
		delete(c.updatesBuffer, seq)
		c.expectedSeq = seq + 1
	}
	if c.isInTurn() {
		log.Printf("%s is in turn\n", c.Name())
		c.botTurn <- struct{}{}
	}
}

func (c *BotClient) isInTurn() bool {
	if currentPlayers, err := c.Room().GameCurrentPlayersClients(); err == nil {
		for _, clientInTurn := range currentPlayers {
			if clientInTurn == c {
				return true
			}
		}
	}
	return false
}

// Incoming returns bot's incoming channel
func (c *BotClient) Incoming() chan []byte {
	return c.incoming
}

// Name returns bot's name
func (c *BotClient) Name() string {
	return c.name
}

// SetName sets bot's name
func (c *BotClient) SetName(v string) interfaces.Client {
	c.name = v
	return c
}

// Close sends a quitting signal that will end the ReadPump() and WritePump()
// goroutines of this instance
func (c *BotClient) Close() {
	close(c.endReadPump)
	close(c.endWritePump)
}

// IsBot returns true because this client is managed by a bot
func (c *BotClient) IsBot() bool {
	return true
}

// Room returns the room where the bot client is in
func (c *BotClient) Room() interfaces.Room {
	return c.room
}

// SetRoom sets the bot client's room
func (c *BotClient) SetRoom(r interfaces.Room) {
	c.room = r
}

// SetTimer is not needed in BotClient
func (c *BotClient) SetTimer(*time.Timer) {
}

// StopTimer is not needed in BotClient
func (c *BotClient) StopTimer() {
}

// StartTimer is not needed in BotClient
func (c *BotClient) StartTimer(d time.Duration) {
}
