package room

import (
	"encoding/json"

	"github.com/svera/sackson-server/interfaces"
)

func newMessage(typeMessage string, args ...interface{}) []byte {
	var encoded []byte

	switch typeMessage {
	case interfaces.TypeMessageError:
		msg := &interfaces.MessageError{
			Type:    "err",
			Content: args[0].(string),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageCurrentPlayers:
		msg := interfaces.MessageCurrentPlayers{
			Type:   interfaces.TypeMessageCurrentPlayers,
			Values: args[0].([]interfaces.MessagePlayer),
		}
		encoded, _ = json.Marshal(msg)

	case interfaces.TypeMessageRoomDestroyed:
		msg := interfaces.MessageRoomDestroyed{
			Type:   interfaces.TypeMessageRoomDestroyed,
			Reason: args[0].(string),
		}
		encoded, _ = json.Marshal(msg)
	}

	return encoded
}
