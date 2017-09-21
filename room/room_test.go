package room

import (
	"testing"

	"encoding/json"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/mocks"
	"github.com/svera/sackson-server/observer"
)

var gamePanickedTriggered int

func setup() (c interfaces.Client, b *mocks.Driver, r *Room) {
	obs := observer.New()
	obs.On(events.GameStarted, func(...interface{}) {})
	obs.On(events.ClientOut, func(...interface{}) {})
	obs.On(events.ClientJoined, func(...interface{}) {})
	obs.On(events.GameStatusUpdated, func(...interface{}) {})
	obs.On(events.ClientsUpdated, func(...interface{}) {})
	obs.On(events.Error, func(...interface{}) {})

	c = &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	b = &mocks.Driver{
		FakeAI: &mocks.AI{FakeIsInTurn: false},
		Calls:  make(map[string]int),
	}

	r = New("test", b, c, make(chan *interfaces.IncomingMessage), make(chan interfaces.Client), &config.Config{Timeout: 1}, obs)
	return c, b, r
}

func TestStartGame(t *testing.T) {
	c, b, r := setup()

	data := []byte(`{"pto": 0}`)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeStartGame,
			Params: (json.RawMessage)(data),
		},
	}
	r.clients[0] = c
	r.Parse(m)

	if b.Calls["StartGame"] != 1 {
		t.Errorf("Room must have StartGame() 1 time, got %d", b.Calls["StartGame"])
	}
}

func TestAddBot(t *testing.T) {
	c, _, r := setup()

	data := []byte(`{"lvl": "chaotic"}`)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeAddBot,
			Params: (json.RawMessage)(data),
		},
	}
	r.Parse(m)

	if len(r.clients) != 1 {
		t.Errorf("Room must have 1 client, got %d", len(r.clients))
	}
}

func TestKickPlayer(t *testing.T) {
	c, _, r := setup()

	data := []byte(`{"ply": 0}`)
	toBeKicked := &mocks.Client{FakeIncoming: make(chan []byte, 2)}

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeKickPlayer,
			Params: (json.RawMessage)(data),
		},
	}

	r.clients[0] = toBeKicked
	r.Parse(m)

	if len(r.clients) != 0 {
		t.Errorf("Room must have no clients after being kicked, got %d", len(r.clients))
	}
}

func TestKickOwnerNotAllowed(t *testing.T) {
	c, _, r := setup()

	data := []byte(`{"ply": 0}`)

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeKickPlayer,
			Params: (json.RawMessage)(data),
		},
	}

	r.clients[0] = c
	r.owner = c
	r.Parse(m)

	if len(r.clients) != 1 {
		t.Errorf("Room must still have owner after trying to kick him/her, got %d", len(r.clients))
	}
}

func TestPlayerQuits(t *testing.T) {
	c, _, r := setup()

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type: interfaces.ControlMessageTypePlayerQuits,
		},
	}

	r.clients[0] = c
	r.Parse(m)

	if len(r.clients) != 0 {
		t.Errorf("Room must have no clients after quitting, got %d", len(r.clients))
	}
}

func TestAddHuman(t *testing.T) {
	c, _, r := setup()

	r.AddHuman(c)

	if len(r.clients) != 1 {
		t.Errorf("Room must have 1 client, got %d", len(r.clients))
	}
}
