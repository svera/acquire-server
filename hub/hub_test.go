package hub

import (
	"testing"
	"time"

	"github.com/olebedev/emitter"
	"github.com/svera/sackson-server/config"
	"github.com/svera/sackson-server/mocks"
)

/* THIS SHOULD BE MOVED TO THE ROOOM TESTS
func TestRunStopsAfterXMinutes(t *testing.T) {
	callbackCalled := false
	var h *Hub
	h = New(&config.Config{Timeout: 0})

	h.Run()
	if !callbackCalled {
		t.Errorf("Hub must stop running and call selfDestructCallBack")
	}
	if !h.wasClosedByTimeout {
		t.Errorf("hub.wasClosedByTimeout must be true")
	}

}
*/
func TestRegister(t *testing.T) {
	var h *Hub
	e := &emitter.Emitter{}
	h = New(&config.Config{Timeout: 1}, e)

	go h.Run()
	c := &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	go c.WritePump()
	h.Register <- c
	if len(h.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
}

func TestUnregister(t *testing.T) {
	var h *Hub
	e := &emitter.Emitter{}
	h = New(&config.Config{Timeout: 1}, e)

	go h.Run()
	c := &mocks.Client{FakeIncoming: make(chan []byte, 2)}
	go c.WritePump()
	h.Register <- c
	h.Unregister <- c
	time.Sleep(time.Second * 1)
	if len(h.clients) != 0 {
		t.Errorf("Hub must have no clients connected after removing it, got %d", len(h.clients))
	}
}
