# Sackson Server

A websocket-driven server written in Go, which allows to play games that implement a bridge interface through message passing between
connected clients and game's logic.

## Architecture

![Sackson server architecture](sackson_server_architecture.png)

## Messages

As stated at the beginning, Sackson server communicates with its different elements (game bridges, clients and hubs) through
message passing. Those messages are in JSON format, and can be categorized between common messages and game-specific ones.

### Common messages

These messages describe server-wide operations, like game creation/destroying, client connection/disconnection, etc, which are
used in all games.
