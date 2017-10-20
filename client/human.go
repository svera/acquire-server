package client

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/interfaces"
)

var (
	mutex sync.RWMutex
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
	quit     chan struct{}
	game     string
}

// NewHuman returns a new Human instance
func NewHuman(w http.ResponseWriter, r *http.Request, cfg *config.Config) (interfaces.Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
		CheckOrigin: func(r *http.Request) bool {
			if r.Header.Get("Origin") != cfg.AllowedOrigin && cfg.AllowedOrigin != "*" {
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
		quit:     make(chan struct{}),
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
		case <-c.quit:
			return
		}
	}
}

// Room returns the room where the client is in
func (c *Human) Room() interfaces.Room {
	mutex.Lock()
	defer mutex.Unlock()
	return c.room
}

// SetRoom sets the client's room
func (c *Human) SetRoom(r interfaces.Room) {
	mutex.Lock()
	defer mutex.Unlock()
	c.room = r
}

// Incoming returns the client's incoming channel
func (c *Human) Incoming() chan []byte {
	mutex.Lock()
	defer mutex.Unlock()
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
	mutex.Lock()
	defer mutex.Unlock()
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}

// Close closes connection through the websocket
func (c *Human) Close() {
	c.write(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	close(c.quit)
}

// IsBot returns false because this client is not managed by a bot
func (c *Human) IsBot() bool {
	return false
}

// SetTimer sets client's timer, which manages when to expel a client from a room due to inactivity
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

// SetGame specifies the name of the game the human client is going to use
func (c *Human) SetGame(game string) {
	c.game = game
}

// Game specifies the name of the game the bohumant client is using
func (c *Human) Game() string {
	return c.game
}
