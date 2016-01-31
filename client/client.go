package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/svera/acquire/interfaces"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

type Client struct {
	Pl       interfaces.Player
	ws       *websocket.Conn
	Incoming chan []byte // Channel storing incoming messages
}

func New(w http.ResponseWriter, r *http.Request, pl interfaces.Player) (*Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		Pl:       pl,
		Incoming: make(chan []byte, maxMessageSize),
		ws:       ws,
	}, nil
}

// ReadPump reads input from the user and writes it to the passed channel
func (c *Client) ReadPump(channel chan *Message, unregister chan *Client) {
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
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Incoming:
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

func (c *Client) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}
