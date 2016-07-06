# Sackson Server

A websocket-driven server written in Go, which allows to play games that implement a bridge interface through message passing between
connected clients and game's logic.

## Architecture

![Sackson server architecture](sackson_server_architecture.png)

## Messages

As stated at the beginning, Sackson server communicates with its different elements (game bridges, clients and rooms) through
message passing, managed by a structure called Hub. Those messages are in JSON format, and can be categorized between hub messages,
room messages and game-specific ones.

The message flow goes outside in: First the hub parses the incoming message type, if it should be managed by the hub itself it does it, otherwise the message is passed to the game room the client is currently into. Again, the game room parses the message type and if it is not of a type recognized by the room struct, passes it again, this time to the game bridge.

### Hub level messages

These messages describe server-wide operations, basically game room creation/destroying.

### Room level messages

The following are the actions a room can execute, and therefore the types of messages it can manage:

* Start a game.
* Add a bot to a game.
* Kick a player out of a room.
* Manage a player leaving a room.

### Game level messages

If a message does not fall into any of the above two categories, it is considered to be a game-specific message and thus will be managed by
the room game bridge.
