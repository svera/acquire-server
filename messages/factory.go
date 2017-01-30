package messages

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

// New returns a new message of the passed type encoded in JSON
func New(typeMessage string, args ...interface{}) []byte {
	var encoded []byte

	switch typeMessage {
	case interfaces.TypeMessageError:
		msg := &interfaces.MessageError{
			Type:    interfaces.TypeMessageError,
			Content: args[0].(string),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageCurrentPlayers:
		msg := interfaces.MessageCurrentPlayers{
			Type:   interfaces.TypeMessageCurrentPlayers,
			Values: args[0].([]interfaces.PlayerData),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageClientOut:
		msg := interfaces.MessageClientOut{
			Type:   interfaces.TypeMessageClientOut,
			Reason: args[0].(string),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageRoomCreated:
		msg := interfaces.MessageRoomCreated{
			Type: interfaces.TypeMessageRoomCreated,
			ID:   args[0].(string),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageRoomsList:
		msg := interfaces.MessageRoomsList{
			Type:   interfaces.TypeMessageRoomsList,
			Values: args[0].([]string),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageJoinedRoom:
		msg := interfaces.MessageJoinedRoom{
			Type:         interfaces.TypeMessageJoinedRoom,
			ClientNumber: args[0].(int),
		}
		encoded, _ = json.Marshal(msg)

	}
	return encoded
}
