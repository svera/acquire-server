$(function() {
    if (!window["WebSocket"]) {
        return;
    }

  $("#newGame").submit(function(e) {
    e.preventDefault();

    $.post("/create",
        {game: "acquire"},
        function(data, status) {
            window.location.replace("/"+data);
        });
  });

});
