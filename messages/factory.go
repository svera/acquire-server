package messages

import "github.com/svera/sackson-server/interfaces"
import "encoding/json"

// New returns a new message of the passed type encoded in JSON
func New(typeMessage string, args ...interface{}) interface{} {
	var msg interface{}

	switch typeMessage {
	case interfaces.TypeMessageError:
		msg = &interfaces.MessageError{
			Content: args[0].(string),
		}

	case interfaces.TypeMessageCurrentPlayers:
		msg = interfaces.MessageCurrentPlayers{
			Values: args[0].(map[string]interfaces.PlayerData),
		}

	case interfaces.TypeMessageClientOut:
		msg = interfaces.MessageClientOut{
			Reason: args[0].(string),
		}

	case interfaces.TypeMessageRoomsList:
		msg = interfaces.MessageRoomsList{
			Values: args[0].([]string),
		}

	case interfaces.TypeMessageJoinedRoom:
		msg = interfaces.MessageJoinedRoom{
			ClientNumber: args[0].(int),
			ID:           args[1].(string),
			Owner:        args[2].(bool),
		}

	case interfaces.TypeMessageGameStarted:
		msg = interfaces.MessageGameStarted{
			GameParameters: args[0].(json.RawMessage),
		}

	}
	return msg
}
