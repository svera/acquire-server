$(function() {
    if (!window["WebSocket"]) {
        return;
    }

    var playButton = $("#playButton");
    var conn = new WebSocket('ws://' + window.location.host + '/ws');

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
        conn.send(tl.val());
        console.log(tl.val());
    });
});