package null

import (
	"encoding/json"

	"github.com/svera/tbg-server/client"
	"github.com/svera/tbg-server/interfaces"
)

type NullBridge struct{}

func (s *NullBridge) ParseMessage(t string, content json.RawMessage) ([]byte, error) {
	return []byte{}, nil
}

func (s *NullBridge) CurrentPlayerNumber() (int, error) {
	return 0, nil
}

func (s *NullBridge) Status(n int) ([]byte, error) {
	return []byte{}, nil
}

func (s *NullBridge) AddPlayer() error {
	return nil
}

func (s *NullBridge) AddBot(params interface{}) (interfaces.Client, error) {
	return &client.NullClient{}, nil
}

func (s *NullBridge) StartGame() error {
	return nil
}

func (s *NullBridge) IsGameOver() bool {
	return false
}
