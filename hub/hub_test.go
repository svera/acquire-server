package hub

import (
	"testing"
	"time"

	"github.com/svera/tbg-server/config"
	"github.com/svera/tbg-server/mocks"
)

func TestRunStopsAfterXMinutes(t *testing.T) {
	callbackCalled := false
	var h *Hub
	h = New(&mocks.Bridge{}, func() { callbackCalled = true }, &config.Config{Timeout: 0})

	h.Run()
	if !callbackCalled {
		t.Errorf("Hub must stop running and call selfDestructCallBack")
	}
	if !h.wasClosedByTimeout {
		t.Errorf("hub.wasClosedByTimeout must be true")
	}

}

func TestRegister(t *testing.T) {
	var h *Hub
	h = New(&mocks.Bridge{}, func() { h = nil }, &config.Config{Timeout: 1})

	go h.Run()
	c := &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	go c.WritePump()
	h.Register <- c
	if len(h.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
	close(h.stop)
}

func TestUnregister(t *testing.T) {
	var h *Hub
	h = New(&mocks.Bridge{}, func() { h = nil }, &config.Config{Timeout: 1})

	go h.Run()
	c := &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	go c.WritePump()
	h.Register <- c
	h.Unregister <- c
	time.Sleep(time.Second * 1)
	if len(h.clients) != 0 {
		t.Errorf("Hub must have no clients connected after removing it, got %d", len(h.clients))
	}
	close(h.stop)
}
