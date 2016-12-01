package room

import (
	"testing"

	"encoding/json"

	"time"

	"github.com/olebedev/emitter"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/mocks"
)

var r *Room
var e *emitter.Emitter
var b *mocks.Bridge
var c *mocks.Client

func setup() {
	e = &emitter.Emitter{}
	e.Use("*", emitter.Skip)
	c = &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}
	b = &mocks.Bridge{
		FakeClient: &mocks.Client{FakeIncoming: make(chan []byte, 2)},
		Calls:      make(map[string]int),
	}

	r = New("test", b, c, make(chan *interfaces.IncomingMessage), make(chan interfaces.Client), &config.Config{Timeout: 1}, e, roomParams)
}

func TestStartGame(t *testing.T) {
	setup()

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type: interfaces.ControlMessageTypeStartGame,
		},
	}
	r.clients = append(r.clients, c)
	r.Parse(m)

	if b.Calls["StartGame"] != 1 {
		t.Errorf("Room must have StartGame() 1 time, got %d", b.Calls["StartGame"])
	}
}

func TestAddBot(t *testing.T) {
	setup()

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
	setup()

	data := []byte(`{"ply": 0}`)
	toBeKicked := &mocks.Client{FakeIncoming: make(chan []byte, 2)}

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeKickPlayer,
			Params: (json.RawMessage)(data),
		},
	}

	r.clients = append(r.clients, toBeKicked)
	r.Parse(m)

	if len(r.clients) != 0 {
		t.Errorf("Room must have no clients after being kicked, got %d", len(r.clients))
	}
}

func TestKickOwnerNotAllowed(t *testing.T) {
	setup()

	data := []byte(`{"ply": 0}`)

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeKickPlayer,
			Params: (json.RawMessage)(data),
		},
	}

	r.clients = append(r.clients, c)
	r.owner = c
	r.Parse(m)

	if len(r.clients) != 1 {
		t.Errorf("Room must still have owner after trying to kick him/her, got %d", len(r.clients))
	}
}

func TestPlayerQuits(t *testing.T) {
	setup()

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type: interfaces.ControlMessageTypePlayerQuits,
		},
	}

	r.clients = append(r.clients, c)
	r.Parse(m)

	if len(r.clients) != 0 {
		t.Errorf("Room must have no clients after quitting, got %d", len(r.clients))
	}
}

func TestAddHuman(t *testing.T) {
	setup()

	r.AddHuman(c)

	if len(r.clients) != 1 {
		t.Errorf("Room must have 1 client, got %d", len(r.clients))
	}
}
