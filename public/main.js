$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var playButton = $("#playButton");
    var url = [location.host, location.pathname].join('');
    var conn = new WebSocket('ws://' + url + '/join');

    conn.onclose = function(e) {
        playButton.attr("disabled", true);
    };

    // Whenever we receive a message, update
    conn.onmessage = function(e) {
        msg = JSON.parse(e.data)
        switch (msg.typ) {
            case "ini":
                updateHand(msg.hnd)
                break;
            case "upd":
                console.log("update received")
                console.log(msg)
                updateBoard(msg.brd)
                if (msg.hasOwnProperty("hnd")) {
                    updateHand(msg.hnd)
                }
        }
    };

    playButton.on("click", function() {
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
        $("#hand").html("")
        for (var i = 0; i < hand.length; i++) {
            $("#hand").append(
                '<label class="btn btn-default">'+
                    '<input type="radio" name="tiles" value="'+ hand[i] +'">'+
                    '<span>' + hand[i] +'</span>'+
                '</label>');
        }
    }    
});