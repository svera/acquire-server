package acquirebridge

import (
	"encoding/json"

	"github.com/svera/acquire/bots"
	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/tbg-server/client"
	serverInterfaces "github.com/svera/tbg-server/interfaces"
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
	bot      acquireInterfaces.Bot
}

// NewBot returns a new Bot instance
func NewBotClient(b acquireInterfaces.Bot) serverInterfaces.Client {
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
	var m interface{}
	channel := cnl.(chan *client.Message)
	defer func() {
		unregister <- c
	}()

	for {
		select {
		case parsed := <-c.botTurn:
			c.updateBot(parsed)
			m = c.bot.Play()
			bm := m.(bots.Message)
			msg = c.encodeResponse(bm)
		}
		channel <- msg
	}

}

func (c *BotClient) encodeResponse(m bots.Message) *client.Message {
	var enc *client.Message

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
	}
	return enc
}

func (c *BotClient) updateBot(parsed statusMessage) {
	var hand []bots.HandData
	var corps [7]bots.CorpData
	var playerInfo bots.PlayerData
	var rivalsInfo []bots.PlayerData

	for _, tile := range parsed.Hand {
		hand = append(hand, bots.HandData{
			Coords:   tile.Coords,
			Playable: tile.Playable,
		})
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
		Board:      parsed.Board,
		State:      parsed.State,
		Hand:       hand,
		Corps:      corps,
		TiedCorps:  parsed.TiedCorps,
		PlayerInfo: playerInfo,
		RivalsInfo: rivalsInfo,
		LastTurn:   parsed.LastTurn,
	}
	c.bot.Update(st)
}

func (c *BotClient) encodePlayTile(response bots.PlayTileResponseParams) *client.Message {
	params := playTileMessageParams{
		Tile: response.Tile,
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

func (c *BotClient) encodeFoundCorporation(response bots.NewCorpResponseParams) *client.Message {
	params := newCorpMessageParams{
		Corporation: response.Corporation,
	}
	ser, _ := json.Marshal(params)
	return &client.Message{
		Author: c,
		Content: client.MessageContent{
			Type:   messageTypeFoundCorporation,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeBuyStock(response bots.BuyResponseParams) *client.Message {
	params := buyMessageParams{
		Corporations: response.Corporations,
	}
	ser, _ := json.Marshal(params)
	return &client.Message{
		Author: c,
		Content: client.MessageContent{
			Type:   messageTypeBuyStock,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeSellTrade(response bots.SellTradeResponseParams) *client.Message {
	params := sellTradeMessageParams{
		Corporations: map[string]sellTrade{},
	}
	for k, v := range response.Corporations {
		params.Corporations[k] = sellTrade{v.Sell, v.Trade}
	}
	ser, _ := json.Marshal(params)
	return &client.Message{
		Author: c,
		Content: client.MessageContent{
			Type:   messageTypeSellTrade,
			Params: ser,
		},
	}
}

func (c *BotClient) encodeUntieMerge(response bots.UntieMergeResponseParams) *client.Message {
	params := untieMergeMessageParams{
		Corporation: response.Corporation,
	}
	ser, _ := json.Marshal(params)
	return &client.Message{
		Author: c,
		Content: client.MessageContent{
			Type:   messageTypeUntieMerge,
			Params: ser,
		},
	}
}

// WritePump gets updates from the hub
func (c *BotClient) WritePump() {
	var parsed statusMessage

	for {
		select {
		case message, ok := <-c.incoming:
			if !ok {
				return
			}

			if err := json.Unmarshal(message, &parsed); err == nil {
				if parsed.PlayerInfo.Enabled {
					c.botTurn <- parsed
				}
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
