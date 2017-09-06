package hub

import (
	"encoding/json"
	"log"
	"time"

	"strings"

	"github.com/svera/sackson-server/drivers"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/room"
)

func (h *Hub) createRoomAction(m *interfaces.IncomingMessage) error {
	var parsed interfaces.MessageCreateRoomParams
	var err error
	var driver interfaces.Driver

	if err = json.Unmarshal(m.Content.Params, &parsed); err != nil {
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
	id := h.generateID()
	h.rooms[id] = room.New(id, b, owner, h.Messages, h.Unregister, h.configuration, h.observer)

	timer := time.AfterFunc(time.Second*h.configuration.Timeout, func() {
		if h.configuration.Debug {
			log.Printf("Destroying room %s due to timeout\n", id)
		}
		h.destroyRoom(id, interfaces.ReasonRoomDestroyedTimeout)
	})
	h.rooms[id].SetTimer(timer)

	h.observer.Trigger("messageCreated", h.clients, h.createUpdatedRoomListMessage(), interfaces.TypeMessageRoomsList)

	if h.configuration.Debug {
		log.Printf("Room %s created\n", id)
	}
	h.rooms[id].AddHuman(owner)

	return id
}

func (h *Hub) generateID() string {
	letters := `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	var locator string
	var randomPosition int
	numberLetters := len(letters)
	for {
		for i := 0; i < 5; i++ {
			randomPosition = rn.Intn(numberLetters - 1)
			locator += string(letters[randomPosition])
		}
		if _, exists := h.rooms[locator]; !exists {
			return locator
		}
	}
}
