$(document).ready(function(){
    $("#player-1").jPlayer({
        ready: function () {
        $(this).jPlayer("setMedia", {
            title: "Bubble",
            oga: "http://localhost:8001"
        }).jPlayer("play");
        },
        supplied: "oga",
        volume: 0.3
    });
    $(".result").slice(20).hide();
    conn = new WebSocket("ws:///localhost:8080/ws");
    conn.onclose = function(evt) {
        console.log("WS closed")
    }
    conn.onmessage = function(evt) {
        console.log(evt.data)
    }
});
