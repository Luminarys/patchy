var ctime = 0
var stime = 0
var songProg

$(document).ready(function(){


    //Load now playing
    $.get("/np", function(data) {
        var song = JSON.parse(data)
        console.log(song)
        $("#npArt").attr("src", "/art/" + song["file"].split("/")[0])
        $("#npSong").text(song["Title"])
        $("#npArtist").text(song["Artist"])
        $("#npAlbum").text(song["Album"])
        $("#songTime").text(secToMin(song["Time"]))
        stime = parseInt(song["Time"])
        $("#curTime").text(secToMin(song["ctime"]))
        ctime = parseInt(song["ctime"])
        var cfile = parseInt(song["cfile"])

        //Load init jplayer
        $("#player-1").jPlayer({
            ready: function () {
            $(this).jPlayer("setMedia", {
                mp3: "/queue/ns" + cfile + ".mp3?" + randString()
            }).jPlayer("play", ctime);
            },
            supplied: "mp3",
            preload: "auto",
            volume: 0.3
        });

        console.log("Initialized and started Player1 using file /queue/ns" + cfile + ".mp3")

        $("#songProgress").css("width", (100 * parseInt(song["ctime"])/parseInt(song["Time"])).toString() + "%")
        songProg = window.setInterval(updateSong, 1000);
        $("#cs").attr("val", cfile)
    });

    //Load Library
    $.get("/library", function(data) {
        var songs = JSON.parse(data)
        $.each(songs, function(index, song) {
            $(".search-results").append('<div class="result"><img alt="Album art" src="/art/' + song["file"].split("/")[0] + '"><div><p><strong>' + song["Title"] + '</strong></p><p>by <strong>' + song["Artist"] + '</strong></p><p>from <strong>' + song["Album"] +' </strong></p><button class="btn btn-primary btn-block">Request</button></div></div>')
        }); 
    });

    $(".result").slice(20).hide();

    //Initialize Websocket
    conn = new WebSocket("ws:///localhost:8080/ws");
    conn.onclose = function(evt) {
        console.log("WS closed")
    }
    conn.onmessage = function(evt) {
        console.log(evt.data)
        var cmd = JSON.parse(evt.data)
        //Update now playing
        if(cmd["cmd"] == "done"){
            newSong(cmd)
        }
        //Pause over, start next song
        if(cmd["cmd"] == "NS"){
            startSong()
        }
    }
});

function newSong(song) {
    $("#player-1").jPlayer("stop")

    var cs = $("#cs").attr("val")
    if(cs == 1){
        cs = 2
    }else{
        cs = 1
    }
    $("#cs").attr("val", cs.toString())

    var cf = "/queue/ns" + cs.toString() + ".mp3?" + randString()
    
    $("#player-1").jPlayer("clearMedia")
    $("#player-1").jPlayer("setMedia", {
            mp3: cf
    });
    console.log("Set Player1 to load song /queue/ns" + cs.toString() + ".mp3 in the background")

    window.clearInterval(songProg)

    $("#npArt").attr("src", song["Cover"])
    $("#npSong").text(song["Title"])
    $("#npArtist").text(song["Artist"])
    $("#npAlbum").text(song["Album"])
    $("#songTime").text(secToMin(song["Time"]))
    $("#curTime").text(secToMin("0"))
    $("#songProgress").css("width", "0%")

    stime = parseInt(song["Time"])
    ctime = 0
}

function startSong() {
    $("#player-1").jPlayer("play")
    console.log("Set Player1 to start playing song")
    songProg = window.setInterval(updateSong, 1000);
}

function secToMin(seconds){
    seconds = parseInt(seconds)
    var min = Math.floor(seconds/60);
    var rsecs = seconds - min * 60
    if (rsecs < 10){
        return min.toString() + ":0" + rsecs.toString();
    }else{
        return min.toString() + ":" + rsecs.toString();
    }
}

function updateSong() {
    if(ctime < stime){
        ctime++
        $("#curTime").text(secToMin(ctime))
        $("#songProgress").css("width", (100 * parseInt(ctime)/parseInt(stime)).toString() + "%")
    }

}

function randString() {
    var text = "";
    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    for( var i=0; i < 5; i++ )
        text += possible.charAt(Math.floor(Math.random() * possible.length));

    return text;
}
