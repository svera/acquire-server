package hub

import (
	"encoding/json"
	"log"
	"time"

	"strings"

	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/drivers"
	"github.com/svera/sackson-server/events"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/messages"
	"github.com/svera/sackson-server/room"
)

// NewRoom holds a factory function that can be replaced in tests, so it returns a mocked Room instead
var NewRoom = func(ID string, b interfaces.Driver, owner interfaces.Client, messages chan *interfaces.IncomingMessage, unregister chan interfaces.Client, cfg *config.Config, ob interfaces.Observer) interfaces.Room {
	return room.New(ID, b, owner, messages, unregister, cfg, ob)
}

// GenerateID returns a random string locator
var GenerateID = func() string {
	letters := `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	var locator string
	var randomPosition int
	numberLetters := len(letters)
	for i := 0; i < 5; i++ {
		randomPosition = rn.Intn(numberLetters - 1)
		locator += string(letters[randomPosition])
	}
	return locator
}

func (h *Hub) createRoomAction(m *interfaces.IncomingMessage) error {
	var parsed messages.CreateRoom
	var err error
	var driver interfaces.Driver

	if err = json.Unmarshal(m.Content, &parsed); err != nil {
		return err
	}
	if driver, err = drivers.Create(parsed.DriverName); err != nil {
		return err
	}

	if strings.TrimSpace(parsed.ClientName) != "" {
		m.Author.SetName(parsed.ClientName)
	}
	h.createRoom(driver, m.Author)
	return nil
}

func (h *Hub) createRoom(b interfaces.Driver, owner interfaces.Client) string {
	exists := true
	var ID string
	for exists {
		ID = GenerateID()
		_, exists = h.rooms[ID]
	}

	h.rooms[ID] = NewRoom(ID, b, owner, h.Messages, h.Unregister, h.configuration, h.observer)

	timer := time.AfterFunc(time.Second*h.configuration.Timeout, func() {
		if h.configuration.Debug {
			log.Printf("Destroying room %s due to timeout\n", ID)
		}
		h.destroyRoom(ID, messages.ReasonRoomDestroyedTimeout)
	})
	h.rooms[ID].SetTimer(timer)

	h.observer.Trigger(events.RoomCreated{Room: h.rooms[ID]})

	if h.configuration.Debug {
		log.Printf("Room %s created\n", ID)
	}
	h.rooms[ID].AddHuman(owner)

	return ID
}
