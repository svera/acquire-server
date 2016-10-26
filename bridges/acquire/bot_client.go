package acquirebridge

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/svera/acquire/bots"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	serverInterfaces "github.com/svera/sackson-server/interfaces"
)

const (
	maxMessageSize = 1024 * 1024
)

// BotClient is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type BotClient struct {
	name         string
	incoming     chan []byte // Channel storing incoming messages
	endReadPump  chan bool
	endWritePump chan bool
	botTurn      chan statusMessage
	bot          acquireInterfaces.Bot
	room         serverInterfaces.Room
}

// NewBotClient returns a new Bot instance
func NewBotClient(b acquireInterfaces.Bot, room serverInterfaces.Room) serverInterfaces.Client {
	return &BotClient{
		incoming:     make(chan []byte, maxMessageSize),
		endReadPump:  make(chan bool),
		endWritePump: make(chan bool),
		botTurn:      make(chan statusMessage),
		bot:          b,
		room:         room,
	}
}

// ReadPump listens to the botTurn channel (see the WritePump function) and, when
// an update message comes this way, updates the bot game status information
// and gets its next play, sending it back to the hub
func (c *BotClient) ReadPump(cnl interface{}, unregister chan serverInterfaces.Client) {
	var msg *serverInterfaces.MessageFromClient
	var m interface{}
	channel := cnl.(chan *serverInterfaces.MessageFromClient)
	defer func() {
		unregister <- c
	}()

	for {
		select {
		case <-c.endReadPump:
			return

		case parsed := <-c.botTurn:
			c.updateBot(parsed)
			m = c.bot.Play()
			bm := m.(bots.Message)
			msg = c.encodeResponse(bm)
		}
		channel <- msg
	}

}

func (c *BotClient) encodeResponse(m bots.Message) *serverInterfaces.MessageFromClient {
	var enc *serverInterfaces.MessageFromClient

	switch m.Type {
	case bots.PlayTileResponseType:
		enc = c.encodePlayTile(m.Params.(bots.PlayTileResponseParams))
	case bots.NewCorpResponseType:
		enc = c.encodeFoundCorporation(m.Params.(bots.NewCorpResponseParams))
	case bots.BuyResponseType:
		enc = c.encodeBuyStock(m.Params.(bots.BuyResponseParams))
	case bots.SellTradeResponseType:
		enc = c.encodeSellTrade(m.Params.(bots.SellTradeResponseParams))
	case bots.UntieMergeResponseType:
		enc = c.encodeUntieMerge(m.Params.(bots.UntieMergeResponseParams))
	case bots.EndGameResponseType:
		enc = c.encodeEndGame()
	default:
		panic(fmt.Sprintf("Unrecognized bot response: %s", m.Type))
	}
	return enc
}

func (c *BotClient) updateBot(parsed statusMessage) {
	hand := map[string]bool{}
	var corps [7]bots.CorpData
	var playerInfo bots.PlayerData
	var rivalsInfo []bots.PlayerData

	for coords, playable := range parsed.Hand {
		hand[coords] = playable
	}
	for i := range parsed.Corps {
		corps[i] = bots.CorpData{
			Name:            parsed.Corps[i].Name,
			Price:           parsed.Corps[i].Price,
			MajorityBonus:   parsed.Corps[i].MajorityBonus,
			MinorityBonus:   parsed.Corps[i].MinorityBonus,
			RemainingShares: parsed.Corps[i].RemainingShares,
			Size:            parsed.Corps[i].Size,
			Defunct:         parsed.Corps[i].Defunct,
			Tied:            parsed.Corps[i].Tied,
		}
	}
	playerInfo = bots.PlayerData{
		Cash:        parsed.PlayerInfo.Cash,
		OwnedShares: parsed.PlayerInfo.OwnedShares,
	}
	for _, rival := range parsed.RivalsInfo {
		rivalsInfo = append(rivalsInfo, bots.PlayerData{
			Cash:        rival.Cash,
			OwnedShares: rival.OwnedShares,
		})
	}

	st := bots.Status{
		Board:       parsed.Board,
		State:       parsed.State,
		Hand:        hand,
		Corps:       corps,
		PlayerInfo:  playerInfo,
		RivalsInfo:  rivalsInfo,
		IsLastRound: parsed.IsLastRound,
	}
	c.bot.Update(st)
}

func (c *BotClient) encodePlayTile(response bots.PlayTileResponseParams) *serverInterfaces.MessageFromClient {
	params := playTileMessageParams{
		Tile: response.Tile,
	}
	ser, _ := json.Marshal(params)
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type:   messageTypePlayTile,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeFoundCorporation(response bots.NewCorpResponseParams) *serverInterfaces.MessageFromClient {
	params := newCorpMessageParams{
		CorporationIndex: response.CorporationIndex,
	}
	ser, _ := json.Marshal(params)
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type:   messageTypeFoundCorporation,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeBuyStock(response bots.BuyResponseParams) *serverInterfaces.MessageFromClient {
	params := buyMessageParams{
		CorporationsIndexes: response.CorporationsIndexes,
	}
	ser, _ := json.Marshal(params)
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type:   messageTypeBuyStock,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeSellTrade(response bots.SellTradeResponseParams) *serverInterfaces.MessageFromClient {
	params := sellTradeMessageParams{
		CorporationsIndexes: map[string]sellTrade{},
	}
	for k, v := range response.CorporationsIndexes {
		params.CorporationsIndexes[k] = sellTrade{v.Sell, v.Trade}
	}
	ser, _ := json.Marshal(params)
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type:   messageTypeSellTrade,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeUntieMerge(response bots.UntieMergeResponseParams) *serverInterfaces.MessageFromClient {
	params := untieMergeMessageParams{
		CorporationIndex: response.CorporationIndex,
	}
	ser, _ := json.Marshal(params)
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type:   messageTypeUntieMerge,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeEndGame() *serverInterfaces.MessageFromClient {
	return &serverInterfaces.MessageFromClient{
		Author: c,
		Content: serverInterfaces.MessageFromClientContent{
			Type: messageTypeEndGame,
		},
	}
}

// WritePump gets updates from the hub
func (c *BotClient) WritePump() {
	for {
		select {
		case <-c.endWritePump:
			return

		case message, ok := <-c.incoming:
			var parsed statusMessage
			if !ok {
				return
			}
			if err := json.Unmarshal(message, &parsed); err == nil {
				if parsed.PlayerInfo.InTurn {
					c.botTurn <- parsed
				}
			}
		}
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
func (c *BotClient) SetName(v string) serverInterfaces.Client {
	c.name = v
	return c
}

// Close sends a quitting signal that will end the ReadPump() and WritePump()
// goroutines of this instance
func (c *BotClient) Close() {
	c.endReadPump <- true
	c.endWritePump <- true
}

// IsBot returns true because this client is managed by a bot
func (c *BotClient) IsBot() bool {
	return true
}

// Room returns the room where the bot client is in
func (c *BotClient) Room() serverInterfaces.Room {
	return c.room
}

// SetRoom sets the bot client's room
func (c *BotClient) SetRoom(r serverInterfaces.Room) {
	c.room = r
}

// SetTimer is not needed in BotCLient
func (c *BotClient) SetTimer(*time.Timer) {
}

func (c *BotClient) StopTimer() {
}

func (c *BotClient) StartTimer(d time.Duration) {
}
