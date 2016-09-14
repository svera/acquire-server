package bridges

import (
	"errors"

	"github.com/svera/sackson-server/bridges/acquire"
	"github.com/svera/sackson-server/interfaces"
)

// Error messages returned from bridge factory
const (
	BridgeNotFound = "bridge_not_found"
)

// Create returns a new instance of the bridge struct specified
func Create(name string) (interfaces.Bridge, error) {
	switch name {
	case "acquire":
		return acquirebridge.New(), nil
	default:
		return nil, errors.New(BridgeNotFound)
	}
}
