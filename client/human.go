package client

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/svera/tbg-server/interfaces"
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
	owner    bool
}

// NewHuman returns a new Human instance
func NewHuman(w http.ResponseWriter, r *http.Request) (interfaces.Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
		CheckOrigin: func(r *http.Request) bool {
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
	channel := cnl.(chan *Message)
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

		cnt := MessageContent{}
		if err := json.Unmarshal(message, &cnt); err == nil {
			msg := &Message{
				Author:  c,
				Content: cnt,
			}

			channel <- msg
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

func (c *Human) Incoming() chan []byte {
	return c.incoming
}

func (c *Human) Owner() bool {
	return c.owner
}

func (c *Human) SetOwner(v bool) interfaces.Client {
	c.owner = v
	return c
}

func (c *Human) Name() string {
	return c.name
}

func (c *Human) SetName(v string) interfaces.Client {
	c.name = v
	return c
}

func (c *Human) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}
