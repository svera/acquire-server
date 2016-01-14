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

    // Whenever we receive a message, update textarea
    conn.onmessage = function(e) {
        tiles = JSON.parse(e.data)
        buttons = $(".tl")
        $.each(tiles, function(index, t){
            $(buttons[index]).find("input").attr("value", t)
            $(buttons[index]).find("span").text(t)
        });
        if (e.data != content.val()) {
            content.val(e.data);
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
});