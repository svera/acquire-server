package hub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/mocks"
	"github.com/svera/sackson-server/observer"
)

var b *mocks.Driver

func init() {
	b = &mocks.Driver{}
}

func setup() (h *Hub, c interfaces.Client) {
	h = New(&config.Config{Timeout: 5, Debug: true}, observer.New())
	c = &mocks.Client{FakeIncoming: make(chan []byte, 2), FakeName: "TestClient", FakeGame: "test"}
	return h, c
}

func TestRegister(t *testing.T) {
	h, c := setup()
	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	if len(h.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
}

func TestUnregister(t *testing.T) {
	h, c := setup()
	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	h.Unregister <- c
	time.Sleep(time.Millisecond * 100)
	if len(h.clients["test"]) != 0 {
		t.Errorf("Hub must have no clients connected after removing it, got %d", len(h.clients))
	}
}

func TestCreateRoom(t *testing.T) {
	h, c := setup()
	go h.Run()

	go c.WritePump()
	h.Register <- c

	data := []byte(`{"drv": "test"}`)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeCreateRoom,
			Params: (json.RawMessage)(data),
		},
	}
	h.Messages <- m
	// We add a little pause to let the hub process the incoming message, as it does it concurrently
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 1 {
		t.Errorf("Hub must have 1 room, got %d", len(h.rooms))
	}
}

func TestDestroyRoom(t *testing.T) {
	h, c := setup()
	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	h.createRoom(b, c)
	time.Sleep(time.Millisecond * 100)
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

func TestDestroyRoomAfterXSeconds(t *testing.T) {
	h, c := setup()
	h.configuration.Timeout = 1
	go h.Run()

	go c.WritePump()
	h.Register <- c

	h.createRoom(b, c)
	time.Sleep(time.Millisecond * 1100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}

func TestDestroyRoomWhenNoHumanClients(t *testing.T) {
	h, c := setup()
	c.(*mocks.Client).FakeIsBot = false
	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	h.createRoom(b, c)
	time.Sleep(time.Millisecond * 100)
	h.Unregister <- c
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}

func TestJoinRoom(t *testing.T) {
	h, c := setup()
	c2 := &mocks.Client{FakeIncoming: make(chan []byte), FakeName: "TestClient2"}

	go h.Run()

	go c.WritePump()
	go c2.WritePump()
	h.Register <- c
	h.Register <- c2

	id := h.createRoom(b, c)
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

/*
func ExampleRecoveredGameDriverPanic() {
	h, c := setup()

	h.createRoom(b, c)
	b.FakeExecute = func(clientName string, t string, content json.RawMessage) error {
		panic("A panic")
	}

	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type: interfaces.ControlMessageTypeAddBot,
		},
	}

	h.Messages <- m

	// Output:
	// Panic in room 'test': A panic
}

func TestUnregisterWhenDestroyingRoom(t *testing.T) {
	h, c := setup()
	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	h.createRoom(b, c)
	time.Sleep(time.Millisecond * 100)
	m := &interfaces.IncomingMessage{
		Author: c,
		Content: interfaces.IncomingMessageContent{
			Type:   interfaces.ControlMessageTypeTerminateRoom,
			Params: json.RawMessage{},
		},
	}
	go func() {
		h.Messages <- m
	}()
	h.Unregister <- c
	time.Sleep(time.Millisecond * 100)
}
*/
