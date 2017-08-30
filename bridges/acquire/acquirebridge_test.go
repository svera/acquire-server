package acquirebridge

import (
	"encoding/json"
	"testing"
)

func TestParseNonExistingTypeMessage(t *testing.T) {
	bridge := New()
	err := bridge.Execute("Test client", "err", json.RawMessage{})
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a non-existing message type")
	}
}

func TestParseWrongTypeMessage(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2", 2: "test3"}

	bridge.StartGame(playerNames)
	data := []byte(`{"aaa": "bbb"}`)
	raw := (json.RawMessage)(data)

	err := bridge.Execute("Test client", "ply", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}

	err = bridge.Execute("Test client", "ncp", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}

	err = bridge.Execute("Test client", "buy", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}

	err = bridge.Execute("Test client", "sel", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}

	err = bridge.Execute("Test client", "unt", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}

	err = bridge.Execute("Test client", "end", raw)
	if err == nil {
		t.Errorf("Bridge must return an error when receiving a malformed message")
	}
}

func TestCurrentPlayerNumbersWithoutGameStarted(t *testing.T) {
	bridge := New()
	if _, err := bridge.CurrentPlayersNumbers(); err == nil {
		t.Errorf("Bridge must return an error when trying to get the current players numbers without a game started")
	}
}

func TestCurrentPlayerNumbersWithGameStarted(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2", 2: "test3"}

	bridge.StartGame(playerNames)

	if _, err := bridge.CurrentPlayersNumbers(); err != nil {
		t.Errorf("Bridge must not return an error when trying to get the current players numbers of a started game")
	}
}

func TestStartGameWithNotEnoughPlayers(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2"}

	if err := bridge.StartGame(playerNames); err == nil {
		t.Errorf("Bridge must return an error when trying to start a game with not enough players")
	}
}

func TestStartGameWithEnoughPlayers(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2", 2: "test3"}

	if err := bridge.StartGame(playerNames); err != nil {
		t.Errorf("Bridge must not return an error when trying to start a game with enough players")
	}
}

func TestStatusWithGameStarted(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2", 2: "test3"}

	bridge.StartGame(playerNames)

	if _, err := bridge.Status(0); err != nil {
		t.Errorf("Bridge must not return an error when trying to get the status of a started game")
	}
}

func TestStatusWithGameNotStarted(t *testing.T) {
	bridge := New()
	if _, err := bridge.Status(0); err == nil {
		t.Errorf("Bridge must return an error when trying to get the status of a non started game")
	}
}

func TestStatusForInexistentPlayer(t *testing.T) {
	bridge := New()
	playerNames := map[int]string{0: "test1", 1: "test2", 2: "test3"}

	bridge.StartGame(playerNames)
	if _, err := bridge.Status(9); err == nil {
		t.Errorf("Bridge must return an error when trying to get the game status of an inexistent player")
	}
}
