var ctime = 0
var stime = 0
var songProg
//Player currently in use
var cPlayer = 1

$(document).ready(function(){
    //Load streaming jplayer
    $("#player-1").jPlayer({
        ready: function () {
        $(this).jPlayer("setMedia", {
            mp3: "http://localhost:8001"
        }).jPlayer("play");
        },
        supplied: "mp3",
        volume: 0.3
    });
    console.log("Initialized Player1 using MPD stream")


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
        if(cfile == 1){
            var nfile = 2
        }else{
            var nfile = 1
        }
        $("#songProgress").css("width", (100 * parseInt(song["ctime"])/parseInt(song["Time"])).toString() + "%")
        songProg = window.setInterval(updateSong, 1000);

        $("#player-2").jPlayer({
            ready: function () {
            $(this).jPlayer("setMedia", {
                mp3: "/queue/ns" + nfile + ".mp3"
            });
            },
            supplied: "mp3",
            volume: 0.3,
            preload: "auto"
        });
        console.log("Initialized Player2 into the background using file /queue/ns" + nfile + ".mp3")
        $("#ns").attr("val", nfile)
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
    $("#player-2").jPlayer("stop")
    var ns = $("#ns").attr("val")
    if(ns == 1){
        ns = 2
    }else{
        ns = 1
    }
    $("#ns").attr("val", ns.toString())
    
    if(cPlayer == 1){
        $("#player-1").jPlayer("setMedia", {
                mp3: "/queue/ns" + ns.toString() + ".mp3"
        });
        console.log("Set Player1 to load song /queue/ns" + ns.toString() + ".mp3 in the background")
    }else{
        $("#player-2").jPlayer("setMedia", {
                mp3: "/queue/ns" + ns.toString() + ".mp3"
        });
        console.log("Set Player2 to load song /queue/ns" + ns.toString() + ".mp3 in the background")
    }

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
    if(cPlayer == 1){
        $("#player-2").jPlayer("play")
        cPlayer = 2
        console.log("Set Player2 to start playing song")
    }else{
        $("#player-1").jPlayer("play")
        cPlayer = 1
        console.log("Set Player1 to start playing song")
    }
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
    if(ctime <= stime){
        ctime++
        $("#curTime").text(secToMin(ctime))
        $("#songProgress").css("width", (100 * parseInt(ctime)/parseInt(stime)).toString() + "%")
    }

}
