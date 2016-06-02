package hub

import (
	"testing"

	"github.com/svera/tbg-server/interfaces"
	"github.com/svera/tbg-server/mocks"
)

func TestRunStopsAfterXMinutes(t *testing.T) {
	callbackCalled := false
	hub := &Hub{
		stop:                 make(chan struct{}),
		selfDestructCallBack: func() { callbackCalled = true },
		timeout:              0,
		wasClosedByTimeout:   false,
	}

	hub.Run()
	if !callbackCalled {
		t.Errorf("Hub must stop running and call selfDestructCallBack")
	}
	if !hub.wasClosedByTimeout {
		t.Errorf("hub.wasClosedByTimeout must be true")
	}

}

func TestRegister(t *testing.T) {
	hub := &Hub{
		stop:       make(chan struct{}),
		Register:   make(chan interfaces.Client),
		timeout:    1,
		clients:    []interfaces.Client{},
		gameBridge: &mocks.Bridge{},
	}

	go hub.Run()
	c := &mocks.Client{FakeIncoming: make(chan []byte)}
	go c.WritePump()
	hub.Register <- c
	if len(hub.clients) != 1 {
		t.Errorf("Hub must have 1 client connected after adding it")
	}
	close(hub.stop)
}
