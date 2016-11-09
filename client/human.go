package client

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

// Human is a struct that implements the client interface,
// storing data related to a specific user and provides
// several functions to send/receive data to/from a client using a websocket
// connection
type Human struct {
	name     string
	ws       *websocket.Conn
	incoming chan []byte // Channel storing incoming messages
	room     interfaces.Room
	timer    *time.Timer
}

// NewHuman returns a new Human instance
func NewHuman(w http.ResponseWriter, r *http.Request, cfg *config.Config) (interfaces.Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
		CheckOrigin: func(r *http.Request) bool {
			if r.Header.Get("Origin") != cfg.AllowedOrigin && cfg.AllowedOrigin != "*" {
				http.Error(w, "Origin not allowed", 403)
				return false
			}
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return &Human{}, err
	}

	return &Human{
		incoming: make(chan []byte, maxMessageSize),
		ws:       ws,
	}, nil
}

// ReadPump reads input from the user and writes it to the passed channel
func (c *Human) ReadPump(cnl interface{}, unregister chan interfaces.Client) {
	channel := cnl.(chan *interfaces.IncomingMessage)
	defer func() {
		unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		cnt := interfaces.IncomingMessageContent{}
		if err := json.Unmarshal(message, &cnt); err == nil {
			msg := &interfaces.IncomingMessage{
				Author:  c,
				Content: cnt,
			}

			channel <- msg
		} else {
			log.Println("error decoding message content")
		}
	}
}

// WritePump sends data to the user
func (c *Human) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.incoming:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Room returns the room where the client is in
func (c *Human) Room() interfaces.Room {
	return c.room
}

// SetRoom sets the client's room
func (c *Human) SetRoom(r interfaces.Room) {
	c.room = r
}

// Incoming returns the client's incoming channel
func (c *Human) Incoming() chan []byte {
	return c.incoming
}

// Name returns the client's name
func (c *Human) Name() string {
	return c.name
}

// SetName sets a name for the client
func (c *Human) SetName(v string) interfaces.Client {
	c.name = v
	return c
}

func (c *Human) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}

// Close closes connection through the websocket
func (c *Human) Close() {
	c.write(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.ws.Close()
}

// IsBot returns false because this client is not managed by a bot
func (c *Human) IsBot() bool {
	return false
}

// SetTimer sets room's timer, which manages when to close a room
func (c *Human) SetTimer(t *time.Timer) {
	c.timer = t
	c.StopTimer()
}

// StopTimer stops the client's timer
func (c *Human) StopTimer() {
	if c.timer != nil {
		c.timer.Stop()
	}
}

// StartTimer starts the client's timer
func (c *Human) StartTimer(d time.Duration) {
	c.timer.Reset(d)
}
