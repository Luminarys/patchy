$(document).ready(function(){
    //Load jplayer
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
    //Load Library
    $.get("/library", function(data) {
        var songs = JSON.parse(data)
        $.each(songs, function(index, song) {
            $(".search-results").append('<div class="result"><img alt="Album art" src="/art/' + song["file"].split("/")[0] + '"><div><p><strong>' + song["Title"] + '</strong></p><p>by <strong>' + song["Artist"] + '</strong></p><p>from <strong>' + song["Album"] +' </strong></p><button class="btn btn-primary btn-block">Request</button></div></div>')
        }); 
    });
    $(".result").slice(20).hide();
    conn = new WebSocket("ws:///localhost:8080/ws");
    conn.onclose = function(evt) {
        console.log("WS closed")
    }
    conn.onmessage = function(evt) {
        console.log(evt.data)
        var cmd = JSON.parse(evt.data)
        //Update now playing
        if(cmd["cmd"] == "NP"){
            $("#npArt").attr("src", cmd["Cover"])
            $("#npSong").text(cmd["Title"])
            $("#npArtist").text(cmd["Artist"])
            $("#npAlbum").text(cmd["Album"])
        }
    }
});
