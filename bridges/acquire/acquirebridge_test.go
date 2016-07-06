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
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	bridge.StartGame()
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

func TestCurrentPlayerNumberWithoutGameStarted(t *testing.T) {
	bridge := New()
	if _, err := bridge.CurrentPlayerNumber(); err == nil {
		t.Errorf("Bridge must return an error when trying to get the current player number without a game started")
	}
}

func TestCurrentPlayerNumberWithGameStarted(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	bridge.StartGame()
	if _, err := bridge.CurrentPlayerNumber(); err != nil {
		t.Errorf("Bridge must not return an error when trying to get the current player number of a started game")
	}
}

func TestStartGameWithNotEnoughPlayers(t *testing.T) {
	bridge := New()
	if err := bridge.StartGame(); err == nil {
		t.Errorf("Bridge must return an error when trying to start a game with not enough players")
	}
}

func TestStartGameWithEnoughPlayers(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	if err := bridge.StartGame(); err != nil {
		t.Errorf("Bridge must not return an error when trying to start a game with enough players")
	}
}

func TestStatusWithGameStarted(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	bridge.StartGame()
	if _, err := bridge.Status(0); err != nil {
		t.Errorf("Bridge must not return an error when trying to get the status of a started game")
	}
}

func TestStatusWithGameNotStarted(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	if _, err := bridge.Status(0); err == nil {
		t.Errorf("Bridge must return an error when trying to get the status of a non started game")
	}
}

func TestStatusForInexistentPlayer(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	bridge.StartGame()
	if _, err := bridge.Status(9); err == nil {
		t.Errorf("Bridge must return an error when trying to get the game status of an inexistent player")
	}
}

func TestAddPlayerToAFullGame(t *testing.T) {
	bridge := New()
	for i := 0; i < 7; i++ {
		bridge.AddPlayer("test")
	}
	if err := bridge.AddPlayer("test2"); err == nil {
		t.Errorf("Bridge must return an error when trying to add a player to an already full game")
	}
}

func TestAddPlayerToARunningGame(t *testing.T) {
	bridge := New()
	for i := 0; i < 3; i++ {
		bridge.AddPlayer("test")
	}
	bridge.StartGame()
	if err := bridge.AddPlayer("test2"); err == nil {
		t.Errorf("Bridge must return an error when trying to add a player to a running game")
	}
}
