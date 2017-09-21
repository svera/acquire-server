package interfaces

// ControlMessageTypeCreateRoom defines the value that create room
// messages must have in the Type field.
//
// A MessageJoinedRoom message is sent to the client when the room is create.
//
// A MessageRoomsList message is sent to all clients when the room is created.
//
// The following is a CreateRoom message example:
//   {
//     "typ": "cre",
//     "par": {
//       "bri": "acquire"
//       "nam": "Sergio" // Name of its creator / owner
//     }
//   }
const ControlMessageTypeCreateRoom = "cre"

// ControlMessageTypeJoinRoom defines the value that join room
// messages must have in the Type field.
//
// A MessageJoinedRoom message is sent to the client if he/she joins.
//
// A MessageCurrentPlayers is sent to all clients in the room when the new player joins.
//
// The following is a JoinRoom message example:
//   {
//     "typ": "joi",
//     "par": {
//       "rom": "VWXYZ"
//       "nam": "Miguel"
//     }
//   }
const ControlMessageTypeJoinRoom = "joi"

// ControlMessageTypeTerminateRoom defines the value that terminate room
// messages must have in the Type field.
//
// Can only be issued by the room's owner
//
// A MessageClientOut is sent to all room clients when the room is terminated.
//
// A MessageRoomsList message is sent to all hub clients when the room is terminated.
//
// The following is a TerminateRoom message example:
//   {
//     "typ": "ter",
//     "par": {} // No params needed
//   }
const ControlMessageTypeTerminateRoom = "ter"

// MessageCreateRoomParams defines the needed parameters for a create room
// message.
type MessageCreateRoomParams struct {
	DriverName string `json:"bri"`
	ClientName string `json:"nam"`
}

// MessageJoinRoomParams defines the needed parameters for a join room message.
type MessageJoinRoomParams struct {
	Room       string `json:"rom"`
	ClientName string `json:"nam"`
}
