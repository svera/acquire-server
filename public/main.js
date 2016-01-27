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
        msg = JSON.parse(e.data);
        switch (msg.typ) {
            case "err":
                console.log(msg.cnt);
                break;
            case "upd":
                console.log("update received");
                console.log(msg);
                updateBoard(msg.brd);
                if (msg.ebl) {
                    $("#playButton").attr("disabled", false);
                    playerActive = true;
                } else {
                    $("#playButton").attr("disabled", true);
                    playerActive = false;
                }
                break;
            case "dir":
                if (msg.hasOwnProperty("sta")) {
                    if (msg.sta == "PlayTile") {
                        updateHand(msg.hnd);
                    }
                    if (msg.sta == "FoundCorp") {
                        chooseNewCorporation(msg.ina);
                    }
                    if (msg.sta == "BuyStock") {
                        buyStocks(msg.act);
                    }
                    if (msg.sta == "SellTrade") {
                        sellTrade(msg.sha);
                    }                         
                }
                break;
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
            "par": {
                "til": tl
            }
        };
        return JSON.stringify(message);
    }

    playerControls.on("click", '#newCorpButton', function() {
        corps = $('input[name=corps]:checked')
        conn.send(
            createNewCorpMessage(corps.val())
        );
    });

    createNewCorpMessage = function(corp) {
        var message =  {
            "typ": "ncp",
            "par": {
                "cor": corp
            }
        };
        return JSON.stringify(message);
    }

    playerControls.on("click", '#buyButton', function() {
        var buy = {};
        $('.buyStocks').each(function() {
            buy[this.name] = $(this).val();
        })
        conn.send(
            createNewBuyMessage(buy)
        );        
    });

    createNewBuyMessage = function(buy) {
        var message =  {
            "typ": "buy",
            "par": {}
        };
        for (corp in buy) {
            message["par"][corp] = buy[corp]
        };
        return JSON.stringify(message);
    }

    updateBoard = function(tiles) {
        Object.keys(tiles).forEach(function(key) {
            if (tiles[key] == 'unincorporated') {
                $('#'+key).addClass('unincorporated')
            } else {
                $('#'+key).removeClass('unincorporated')
                $('#'+key).addClass(tiles[key])
            }
        });
    }

    updateHand = function(hand) {
        $("#player-controls").html("")
        var html = '<div class="btn-group" role="group" aria-label="..." data-toggle="buttons">';

        for (var i = 0; i < hand.length; i++) {
            html += '<label class="btn btn-default">'+
                        '<input type="radio" name="tiles" value="'+ hand[i] +'">'+
                        '<span>' + hand[i] +'</span>'+
                    '</label>';
        }
        buttonState = !playerActive ? 'disabled="true"' : '';
        html += '</div>'+
                      '<input type="button" id="playButton" class="btn btn-primary" value="Play tile"' + buttonState +' />';
        $("#player-controls").append(html);
    }

    chooseNewCorporation = function(corporations) {
        $("#player-controls").html("")
        var html = '<div class="btn-group" role="group" aria-label="..." data-toggle="buttons">'+
                        '<p>You have founded a new corporation! Please choose one:</p>';

        for (var i = 0; i < corporations.length; i++) {
            html += '<label class="btn btn-default">'+
                        '<input type="radio" name="corps" value="'+ corporations[i].toLowerCase() +'">'+
                        '<span>' + corporations[i] +'</span>'+
                    '</label>';
        }
        buttonState = !playerActive ? 'disabled="true"' : '';
        html += '</div>'+
                      '<input type="button" id="newCorpButton" class="btn btn-primary" value="Found corporation"' + buttonState +' />'
        $("#player-controls").append(html);
    }

    buyStocks = function(corporations) {
        $("#player-controls").html("");
        var html = '<p>Buy Stocks</p>'+
                    '<ul class="list-unstyled">';
        for (var i = 0; i < corporations.length; i++) {
            html += '<li><label>'+
                        '<span>' + corporations[i] +'</span>'+
                        '<input type="number" min="0" max="3" name="'+ corporations[i].toLowerCase() +'" value="0" class="buyStocks">'+
                    '</label></li>';
        }
        buttonState = !playerActive ? 'disabled="true"' : '';
        html += '</ul>' +
               '<input type="button" id="buyButton" class="btn btn-primary" value="Buy"' + buttonState +' />';
        $("#player-controls").append(html);
    }

    sellTrade = function(corporations) {
        $("#player-controls").html("");
        var html =  '<p>Sell / Trade stock shares</p>'+
                        '<table>'+
                            '<thead><tr>'+
                                '<th>Sell</th><th>Trade</th>'+
                            '</tr></thead>'+
                            '<tbody>';
        for (var i = 0; i < corporations.length; i++) {
            html += '<tr><td>' + corporations[i] +
                        '<input type="number" min="0" name="sell['+ corporations[i].toLowerCase() +']" value="0" class="sell">'+
                    '</td><td>'+
                        '<input type="number" min="0" name="trade['+ corporations[i].toLowerCase() +']" value="0" step="2" class="trade">'+
                    '</td></tr>';
        }      
        buttonState = !playerActive ? 'disabled="true"' : '';                      
        html += '</tbody></table>'+
            '<input type="button" id="buyButton" class="btn btn-primary" value="Sell / Trade"' + buttonState +' />';
        $("#player-controls").append(html);
    }    
});