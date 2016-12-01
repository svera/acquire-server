# Acquire bridge

This package allows to play Acquire games using the Sackson server.

## Acquire accepted messages

These messages are the ones allowed by the Acquire bridge and describe actions performed by players:

* Play a tile.
```
{
  "typ": "ply", // Message type: Play tile
  "par": { // Parameters
    "til": "2A" // Tile coords
  }
}
```

* Found a corporation.
```
{
  "typ": "ncp", // Message type: New corporation
  "par": { // Parameters
    "cor": "2" // Corporation number
  }
}
```

* Buy stock.
```
{
  "typ": "buy", // Message type: Buy stock
  "par": { // Parameters
    "cor": {
        "0": 3,
        "1": 0,
        "2": 2,
        ...
    } 
  }
}
```

* Sell and trade stock.
```
{
  "typ": "sel", // Message type: Sell and trade stock
  "par": { // Parameters
    "cor": {
        "0": {
            "sel": 2,
            "tra": 0
        },
        "1": {
            "sel": 0,
            "tra": 2
        },
        ...
    } 
  }
}
```

* Untie merge.
```
{
  "typ": "unt", // Message type: New corporation
  "par": { // Parameters
    "cor": "2" // Corporation number
  }
}
```

* Claim game ended.
```
{
  "typ": "end" // Message type: End game
}
```

## Acquire updates

Whenever a player performs one of the actions shown above, an update message is issued to all players describing 
the updated game status:

```
{
    "typ": "upd", // Type: update
	"brd": { // Board state
        "1A": "empty",
        "1B": "unincorporated",
        "1C": "empty",
        "1D": "0",
        ...
    }
	"sta": "PlayTile",
    "hnd": {
        "1A": true,
    },
    "cor": [
        {
            "nam": "Hilton",
            "prc": 100,
            "maj": 400,
            "min": 200,
            "rem": 20,
            "siz": 2,
            "def": false,
            "tie": false,        
        },
        ...
    ],
	"ply": {
	    "atv": true, // Is player still in game?
        "nam": "John",
        "trn": true,
        "csh": 6000,
        "own": [
            0: 2,
            1: 0,
            ...
        ]
    },
	"riv": [
        {
            "atv": true, // Is player still in game?
            "nam": "Doe",
            "trn": false,
            "csh": 6000,
            "own": [
                0: 2,
                1: 0,
                ...
            ]
        },
        ...
    ],
	"rnd": 3, // Round number
	"lst": false, // Is last round?
	"his": [ // History log (i18n enabled)
        {
            "key": "translation_key",
            "arg": {
                "argument_name": "argument_value",
                ...
            }
        },
        ...
    ]
}
```