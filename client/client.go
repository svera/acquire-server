package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/svera/acquire/player"
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
	Pl   player.Interface
	ws   *websocket.Conn
	Send chan []byte // Channel storing outcoming messages
}

type ClientMessageContent struct {
	Typ string
	Det map[string]string
}

type ClientMessage struct {
	Author  *Client
	Content ClientMessageContent
}

func New(w http.ResponseWriter, r *http.Request, pl player.Interface) (*Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		Pl:   pl,
		Send: make(chan []byte, maxMessageSize),
		ws:   ws,
	}, nil
}

// ReadPump reads input from the user and writes it to the passed channel
func (c *Client) ReadPump(channel chan *ClientMessage, unregister chan *Client) {
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

		cnt := ClientMessageContent{}
		json.Unmarshal(message, &cnt)

		msg := &ClientMessage{
			Author:  c,
			Content: cnt,
		}

		channel <- msg
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
		case message, ok := <-c.Send:
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
