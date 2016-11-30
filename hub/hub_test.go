package hub

import (
	"testing"
	"time"

	"encoding/json"

	"github.com/olebedev/emitter"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/mocks"
)

/* THIS SHOULD BE MOVED TO THE ROOOM TESTS
func TestRunStopsAfterXMinutes(t *testing.T) {
	callbackCalled := false
	var h *Hub
	h = New(&config.Config{Timeout: 0})

	h.Run()
	if !callbackCalled {
		t.Errorf("Hub must stop running and call selfDestructCallBack")
	}
	if !h.wasClosedByTimeout {
		t.Errorf("hub.wasClosedByTimeout must be true")
	}

}
*/
var h *Hub
var e *emitter.Emitter
var b *mocks.Bridge
var c *mocks.Client

func setup() {
	e = &emitter.Emitter{}
	e.Use("*", emitter.Skip)
	h = New(&config.Config{Timeout: 1}, e)
	go h.Run()

	b = &mocks.Bridge{}
	c = &mocks.Client{FakeIncoming: make(chan []byte, 2)}
}

func TestRegister(t *testing.T) {
	setup()
	go c.WritePump()
	h.Register <- c
	if len(h.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
}

func TestUnregister(t *testing.T) {
	setup()
	go c.WritePump()
	h.Register <- c
	h.Unregister <- c
	// We add a little pause to let the hub process the incoming message, as it does it concurrently
	time.Sleep(time.Millisecond * 100)
	if len(h.clients) != 0 {
		t.Errorf("Hub must have no clients connected after removing it, got %d", len(h.clients))
	}
}

func TestCreateRoom(t *testing.T) {
	setup()
	h.Register <- c

	data := []byte(`{"bri": "acquire", "pto": 0}`)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeCreateRoom,
			Params: (json.RawMessage)(data),
		},
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 1 {
		t.Errorf("Hub must have 1 room, got %d", len(h.rooms))
	}
}

func TestDestroyRoom(t *testing.T) {
	setup()
	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	h.Register <- c

	h.createRoom(b, roomParams, c)

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeTerminateRoom,
			Params: json.RawMessage{},
		},
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}

func TestJoinRoom(t *testing.T) {
	setup()
	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	h.Register <- c

	id := h.createRoom(b, roomParams, c)

	data := []byte(`{"rom": "` + id + `"}`)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeJoinRoom,
			Params: (json.RawMessage)(data),
		},
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms[id].Clients()) != 2 {
		t.Errorf("Hub must have no 2 clients, got %d", len(h.rooms[id].Clients()))
	}
}

/*
func TestDestroyRoomWhenNoHumanClients(t *testing.T) {
	setup()
	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	h.Register <- c

	h.createRoom(b, roomParams, c)
	//c.SetRoom(h.rooms[id])
	time.Sleep(time.Millisecond * 100)
	h.Unregister <- c
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}
*/
