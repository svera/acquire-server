package client

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/svera/sackson-server/interfaces"
)

// BotClient is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type BotClient struct {
	name          string
	incoming      chan []byte // Channel storing incoming messages
	endReadPump   chan struct{}
	endWritePump  chan struct{}
	botTurn       chan struct{}
	ai            interfaces.AI
	room          interfaces.Room
	expectedSeq   int
	updatesBuffer map[int]json.RawMessage
	game          string
}

// NewBot returns a new Bot instance
func NewBot(ai interfaces.AI, room interfaces.Room) interfaces.Client {
	return &BotClient{
		incoming:      make(chan []byte, maxMessageSize),
		endReadPump:   make(chan struct{}),
		endWritePump:  make(chan struct{}),
		botTurn:       make(chan struct{}),
		ai:            ai,
		room:          room,
		expectedSeq:   1,
		updatesBuffer: map[int]json.RawMessage{},
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

// WritePump gets updates from the hub.
// As updates may come in a wrong order, we check if the coming update
// is the one we expect, and if not, store it in an updates buffer until the
// right one comes, then processing all them in the right order.
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

			parsed.Content = json.RawMessage{}
			if err = json.Unmarshal(message, &parsed); err == nil {
				c.updatesBuffer[parsed.SequenceNumber] = parsed.Content
				if parsed.SequenceNumber == c.expectedSeq {
					c.feedPendingUpdatesInOrder()
				}
			}
		}
	}
}

func (c *BotClient) feedPendingUpdatesInOrder() {
	var err error

	for _, seq := range c.getSortedUpdatesBufferKeys() {
		if err = c.ai.FeedGameStatus(c.updatesBuffer[seq]); err == nil {
			if c.ai.IsInTurn() {
				c.botTurn <- struct{}{}
			}
		}
		delete(c.updatesBuffer, seq)
		c.expectedSeq = seq + 1
	}
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

func (c *BotClient) getSortedUpdatesBufferKeys() []int {
	keys := []int{}
	for k := range c.updatesBuffer {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// SetGame specifies the name of the game the bot client is going to use
func (c *BotClient) SetGame(game string) {
	c.game = game
}

// Game returns the name of the game the bot client is using
func (c *BotClient) Game() string {
	return c.game
}
