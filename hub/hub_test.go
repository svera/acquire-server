package hub

import (
	"encoding/json"
	"testing"
	"time"

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
var b *mocks.Bridge

func init() {
	b = &mocks.Bridge{}
}

func setup() (*Hub, interfaces.Client) {
	var h *Hub
	var c *mocks.Client
	var e *emitter.Emitter

	e = &emitter.Emitter{}
	e.Use("*", emitter.Skip)
	h = New(&config.Config{Timeout: 5, Debug: true}, e)
	c = &mocks.Client{FakeIncoming: make(chan []byte, 2), Quit: make(chan struct{})}
	return h, c
}

func TestRegister(t *testing.T) {
	h, c := setup()
	go h.Run()
	defer close(h.Quit)

	go c.WritePump()
	h.Register <- c
	if len(h.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
}

func TestUnregister(t *testing.T) {
	h, c := setup()
	go h.Run()
	defer close(h.Quit)

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
	h, c := setup()
	go h.Run()
	defer close(h.Quit)

	go c.WritePump()
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
	h, c := setup()
	go h.Run()
	defer close(h.Quit)

	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	go c.WritePump()
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

func TestDestroyRoomWhenNoHumanClients(t *testing.T) {
	h, c := setup()
	go h.Run()
	defer close(h.Quit)

	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	go c.WritePump()
	h.Register <- c

	h.createRoom(b, roomParams, c)
	time.Sleep(time.Millisecond * 100)
	h.Unregister <- c
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}

func TestJoinRoom(t *testing.T) {
	h, c := setup()
	c2 := &mocks.Client{FakeIncoming: make(chan []byte, 2), Quit: make(chan struct{})}

	go h.Run()
	defer close(h.Quit)

	roomParams := map[string]interface{}{
		"playerTimeout": time.Duration(0),
	}

	go c.WritePump()
	go c2.WritePump()
	h.Register <- c
	h.Register <- c2

	id := h.createRoom(b, roomParams, c)
	time.Sleep(time.Millisecond * 100)

	data := []byte(`{"rom": "` + id + `"}`)
	m := &interfaces.IncomingMessage{
		Author: c2,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeJoinRoom,
			Params: (json.RawMessage)(data),
		},
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms[id].Clients()) != 2 {
		t.Errorf("Room must have 2 clients, got %d", len(h.rooms[id].Clients()))
	}
}
