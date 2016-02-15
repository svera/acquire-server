package acquirebridge

import (
	"encoding/json"
	"testing"
)

func TestParseWrongMessage(t *testing.T) {
	bridge := New()
	_, err := bridge.ParseMessage("err", json.RawMessage{})
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a non-existing message type")
	}
}
