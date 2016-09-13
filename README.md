# Sackson Server

A websocket-driven server written in Go, which allows to play games that implement a bridge interface through message passing between
connected clients and game's logic.

## Architecture

![Sackson server architecture](sackson_server_architecture.png)

## Messages

As stated at the beginning, Sackson server communicates with its different elements (game bridges, clients and rooms) through
message passing, managed by a structure called Hub. Those messages are in JSON format, and are divided between incoming messages
(from a client to the Hub) and outgoing ones (from the Hub to one or multiple clients).

Incoming messages communicate actions that a user wants to perform and can be
categorized between hub messages, room messages and game-specific ones.

The flow for incoming messages goes outside in: First the hub parses the incoming message type, if it should be managed by the hub itself it does it, otherwise the message is passed to the game room the client is currently into. Again, the game room parses the message type and if it is not of a type recognized by the room struct, passes it again, this time to the game bridge.

On the other hand, outgoing messages are sent to one or multiple clients to communicate events, usually in response to user actions.

#### Hub level messages

These messages describe server-wide operations, basically game room creation/destroying.

* Create a room.
```
{
  "typ": "cre", // Message type: Create room
  "par": { // Parameters
    "bri": "acquire" // Bridge name
  }
}
```

If the room was created, a message with the following format is sent back to the client:
```
{
  "typ": "new", // Message type: New room created
  "id": "abc"   // Room Identifier
}
```

* Destroy a room
```
{
  "typ": "ter", // Message type: Destroy room
  "par": {} // No parameters needed
}
```

If the room was destroyed, a message with the following format is sent back to the client:
```
{
  "typ": "out", // Message type: Room destroyed
  "rea": "abc"   // Reason code, see below
}
```
The reason code can be one of these:
  * `tim`: Room timed out
  * `ncl`: Not enough clients to keep playing
  * `ter`: Room destroyed by owner

##### Rooms list update

Also, when a room is created or destroyed or a game starts, a message listing all available rooms
(rooms which haven't started a game yet) is issued. Its format is as follows:

```
{
  "typ": "rms", // Message type: room list
  "val": ["abcde", "opqrst"...] // Available game rooms IDs
}
```

#### Room level messages

The following are the actions that can be executed in a room, and therefore the types of messages it can manage.
Note that room level messages do not need to specify a room ID because a player
can only be in one room at a time, and the system tracks it:

* Start a game (can only be issued by a room owner):
```
{
  "typ": "ini", // Message type: Init game
  "par": {}
}
```

When a game starts, an _update_ message is broadcast to all clients with the initial status of the game. This _update_ message format depends on the game, look at game bridge documentation for details.

* Add a bot to a game (can only be issued by a room owner):
```
{
  "typ": "bot", // Message type: Add bot
  "par": {
    "lvl": "rookie", // Bot game level
  }
}
```

* Kick a player out of a room (can only be issued by a room owner):
```
{
  "typ": "kck",
  "par": {
    "ply": 1
  }
}
```

* Manage a player leaving a room.
```
{
  "typ": "qui",
  "par": {
    "ply": "slf"
  }
}
```

#### Game level messages

If a message does not fall into any of the above two categories, it is considered to be a game-specific message and thus will be managed by
the room game bridge. Check the bridge documentation for information regarding its messages.
