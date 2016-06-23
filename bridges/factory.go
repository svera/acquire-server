package bridges

import (
	"errors"

	"github.com/svera/sackson-server/bridges/acquire"
	"github.com/svera/sackson-server/interfaces"
)

const (
	BridgeNotFound = "bridge_not_found"
)

func Create(name string) (interfaces.Bridge, error) {
	switch name {
	case "acquire":
		return acquirebridge.New(), nil
	default:
		return nil, errors.New(BridgeNotFound)
	}
}
