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
                buttons = $(".tl")
                $.each(msg.hnd, function(index, t){
                    $(buttons[index]).find("input").attr("value", t)
                    $(buttons[index]).find("span").text(t)
                });
                break;
            case "upd":
                console.log("update received")
                console.log(msg)
                updateBoard(msg.brd)
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
});