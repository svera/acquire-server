package client

import (
	"github.com/gorilla/websocket"
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
	Ws   *websocket.Conn
	Send chan []byte // Channel storing outcoming messages
}

func New(w http.ResponseWriter, r *http.Request) (*Client, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		Send: make(chan []byte, maxMessageSize),
		Ws:   ws,
	}, nil
}

func (c *Client) ReadPump(broadcast chan string, unregister chan *Client) {
	defer func() {
		unregister <- c
		c.Ws.Close()
	}()

	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error {
		c.Ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			break
		}

		broadcast <- string(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Ws.Close()
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
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, message)
}
