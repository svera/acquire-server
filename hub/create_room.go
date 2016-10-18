package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/svera/sackson-server/bridges"
	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/room"
)

func (h *Hub) createRoomAction(m *interfaces.MessageFromClient) {
	var parsed interfaces.MessageCreateRoomParams
	if err := json.Unmarshal(m.Content.Params, &parsed); err == nil {
		if bridge, err := bridges.Create(parsed.BridgeName); err != nil {
			res := &interfaces.MessageError{
				Type:    "err",
				Content: err.Error(),
			}
			response, _ := json.Marshal(res)
			go h.emitter.Emit("messageCreated", []interfaces.Client{m.Author}, response)
		} else {
			roomParams := map[string]interface{}{
				"playerTimeout": parsed.PlayerTimeout,
			}
			id := h.createRoom(bridge, roomParams, m.Author)
			h.rooms[id].AddHuman(m.Author)
		}
	}
}

func (h *Hub) createRoom(b interfaces.Bridge, roomParams map[string]interface{}, owner interfaces.Client) string {
	id := h.generateID()
	h.rooms[id] = room.New(id, b, owner, h.Messages, h.Unregister, h.configuration, h.emitter, roomParams)

	timer := time.AfterFunc(time.Minute*h.configuration.Timeout, func() {
		if h.debug {
			log.Printf("Destroying room %s due to timeout\n", id)
		}
		h.destroyRoom(id, interfaces.ReasonRoomDestroyedTimeout)
	})
	h.rooms[id].SetTimer(timer)

	msgRoomCreated := interfaces.MessageRoomCreated{
		Type: interfaces.TypeMessageRoomCreated,
		ID:   id,
	}
	response, _ := json.Marshal(msgRoomCreated)
	go h.emitter.Emit("messageCreated", []interfaces.Client{owner}, response)

	go h.emitter.Emit("messageCreated", h.clients, h.createUpdatedRoomListMessage())

	if h.debug {
		log.Printf("Room %s created\n", id)
	}

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
