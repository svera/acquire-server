$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var playerControls = $("#player-controls");
    var url = [location.host, location.pathname].join('');
    var conn = new WebSocket('ws://' + url + '/join');
    var playerActive = true;

    conn.onclose = function(e) {
        $("#playButton").attr("disabled", true);
    };

    // Whenever we receive a message, update
    conn.onmessage = function(e) {
        msg = JSON.parse(e.data)
        switch (msg.typ) {
            case "err":
                console.log(msg.cnt)
            case "upd":
                console.log("update received")
                console.log(msg)
                updateBoard(msg.brd)
                if (msg.ebl) {
                    $("#playButton").attr("disabled", false);
                    playerActive = true;
                } else {
                    $("#playButton").attr("disabled", true);
                    playerActive = false;
                }
            case "dir":
                if (msg.hasOwnProperty("hnd")) {
                    updateHand(msg.hnd)
                }
        }
    };

    playerControls.on("click", '#playButton', function() {
        tl = $('input[name=tiles]:checked')
        conn.send(
            createPlayTileMessage(tl.val())
        );
    });

    createPlayTileMessage = function(tl) {
        var message =  {
            "typ": "ply",
            "det": {
                "til": tl
            }
        };
        return JSON.stringify(message);
    }

    updateBoard = function(tiles) {
        Object.keys(tiles).forEach(function(key) {
            if (tiles[key] == 'unincorporated') {
                $('#'+key).addClass('unincorporated')
            }
        });
    }

    updateHand = function(hand) {
        $("#player-controls").html("")
        var html = '<div class="btn-group" role="group" aria-label="..." data-toggle="buttons">'

        for (var i = 0; i < hand.length; i++) {
            html += '<label class="btn btn-default">'+
                        '<input type="radio" name="tiles" value="'+ hand[i] +'">'+
                        '<span>' + hand[i] +'</span>'+
                    '</label>';
        }
        buttonState = !playerActive ? 'disabled="true"' : ''
        html = html + '</div>'+
                      '<input type="button" id="playButton" class="btn btn-primary" value="Play tile"' + buttonState +' />'
        $("#player-controls").append(html)
    }

    chooseNewCorporation = function(corporations) {

    }
});