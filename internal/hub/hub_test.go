package hub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/svera/sackson-server/api"
	"github.com/svera/sackson-server/internal/client"
	"github.com/svera/sackson-server/internal/config"
	"github.com/svera/sackson-server/internal/drivers"
	"github.com/svera/sackson-server/internal/interfaces"
	"github.com/svera/sackson-server/internal/messages"
	"github.com/svera/sackson-server/internal/room"
	"github.com/svera/sackson-server/observer"
)

var b api.Driver

func init() {
	b = drivers.NewMock()
}

func setup() (h *Hub, c *client.Mock) {
	h = New(&config.Config{Timeout: 5, Debug: true}, observer.New())
	c = client.NewMock()
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
	NewRoom = func(ID string, b api.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
		return room.NewMock()
	}
	go h.Run()

	go c.WritePump()
	h.Register <- c

	data := []byte(`{"drv": "test"}`)
	m := &interfaces.IncomingMessage{
		Author:  c,
		Type:    messages.TypeCreateRoom,
		Content: (json.RawMessage)(data),
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
	testRoom := room.NewMock()

	GenerateID = func() string {
		return "testRoom"
	}

	NewRoom = func(ID string, b api.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
		return testRoom
	}

	testRoom.FakeOwner = func() interfaces.Client {
		return c
	}

	c.FakeRoom = func() interfaces.Room {
		return testRoom
	}

	go h.Run()

	go c.WritePump()
	h.Register <- c
	time.Sleep(time.Millisecond * 100)
	h.createRoom(b, c)
	time.Sleep(time.Millisecond * 100)
	m := &interfaces.IncomingMessage{
		Author:  c,
		Type:    messages.TypeTerminateRoom,
		Content: json.RawMessage{},
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if len(h.rooms) != 0 {
		t.Errorf("Hub must have no rooms, got %d", len(h.rooms))
	}
}

func TestDestroyRoomAfterXSeconds(t *testing.T) {
	h, c := setup()
	testRoom := room.NewMock()

	GenerateID = func() string {
		return "testRoom"
	}

	NewRoom = func(ID string, b api.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
		return testRoom
	}

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
	testRoom := room.NewMock()

	GenerateID = func() string {
		return "testRoom"
	}

	NewRoom = func(ID string, b api.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
		return testRoom
	}

	c.FakeIsBot = func() bool {
		return false
	}
	c.FakeRoom = func() interfaces.Room {
		return testRoom
	}
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
	testRoom := room.NewMock()

	NewRoom = func(ID string, b api.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
		return testRoom
	}

	c2 := client.NewMock()

	go h.Run()

	go c.WritePump()
	go c2.WritePump()
	h.Register <- c
	h.Register <- c2

	id := h.createRoom(b, c)
	time.Sleep(time.Millisecond * 100)

	data := []byte(`{"rom": "` + id + `"}`)
	m := &interfaces.IncomingMessage{
		Author:  c2,
		Type:    messages.TypeJoinRoom,
		Content: (json.RawMessage)(data),
	}
	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	if h.rooms[id].(*room.Mock).Calls["AddHuman"] != 2 {
		t.Errorf("Room must have 2 clients, got %d", len(h.rooms[id].Clients()))
	}
}

func ExampleHubRecoversFromRoomPanic() {
	h, c := setup()
	const roomID = "test"

	room := room.NewMock()
	room.FakeGameStarted = func() bool {
		return false
	}
	room.FakeID = func() string {
		return roomID
	}
	room.FakeParse = func(m *interfaces.IncomingMessage) {
		panic("A panic")
	}
	h.rooms[roomID] = room
	c.FakeRoom = func() interfaces.Room {
		return room
	}

	m := &interfaces.IncomingMessage{
		Author:  c,
		Type:    "whatever",
		Content: json.RawMessage{},
	}

	go h.Run()
	h.Register <- c

	h.Messages <- m
	time.Sleep(time.Millisecond * 100)

	// Output:
	// Panic in room 'test': A panic
}
