package messages

// TypeCreateRoom defines the value that create room
// messages must have in the Type field.
//
// A MessageJoinedRoom message is sent to the client when the room is create.
//
// A MessageRoomsList message is sent to all clients when the room is created.
//
// The following is a CreateRoom message example:
//   {
//     "typ": "cre",
//     "cnt": {
//       "drv": "acquire"
//       "nam": "Sergio" // Name of its creator / owner
//     }
//   }
const TypeCreateRoom = "cre"

// CreateRoom defines the needed parameters for a create room
// message.
type CreateRoom struct {
	DriverName string `json:"drv"`
	ClientName string `json:"nam"`
}

// TypeJoinRoom defines the value that join room
// messages must have in the Type field.
//
// A MessageJoinedRoom message is sent to the client if he/she joins.
//
// A MessageCurrentPlayers is sent to all clients in the room when the new player joins.
//
// The following is a JoinRoom message example:
//   {
//     "typ": "joi",
//     "cnt": {
//       "rom": "VWXYZ"
//       "nam": "Miguel"
//     }
//   }
const TypeJoinRoom = "joi"

// JoinRoom defines the needed parameters for a join room message.
type JoinRoom struct {
	Room       string `json:"rom"`
	ClientName string `json:"nam"`
}

// TypeTerminateRoom defines the value that terminate room
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
//     "cnt": {} // No content needed
//   }
const TypeTerminateRoom = "ter"
