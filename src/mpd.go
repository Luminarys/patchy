package main

import (
	"encoding/json"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"os"
	"strconv"
	"time"
)

func handleSongs(utaChan chan string, queue []mpd.Attrs, h *hub) {
	var conn *mpd.Client
	var status mpd.Attrs
	var queuePos int = 0
	var cFile int = 1

	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	defer conn.Close()

	w, err := mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	defer w.Close()

	status, err = conn.Status()

	for {
		select {
		case <-w.Event:
			status, err = conn.Status()
			if err != nil {
				//Connections seem to drop often, so reconnect when this happens
				fmt.Println("Couldn't get current status! Error: " + err.Error())
				conn.Close()

				fmt.Println("Reconnecting...")
				conn, err = mpd.Dial("tcp", "127.0.0.1:6600")
				if err != nil {
					fmt.Println("Error: could not connect to MPD, exiting")
					os.Exit(1)
				}
				defer conn.Close()

				status, err = conn.Status()
			}
			pos, _ := strconv.ParseFloat(status["elapsed"], 64)
			if pos == 0.000 && status["state"] == "play" {
				//Stop us from getting into an infinite loop by waiting 25 ms
				time.Sleep(100 * time.Millisecond)
				song, err := conn.CurrentSong()
				//Prep next song
				queuePos++
				fmt.Println("Next song:")
				fmt.Println(queue[queuePos+1])
				fmt.Println("The file to be replaced with the next song is:" + strconv.Itoa(cFile))
				//If cFile is 1, then the just finished song used ns1, otherwise it was using ns2.mp3
				if cFile == 1 {
					os.Rename("static/queue/next.mp3", "static/queue/ns1.mp3")
					//Transcode next song
					go transcode(musicDir + "/" + queue[queuePos+2]["file"])
					cFile = 2
				} else {
					os.Rename("static/queue/next.mp3", "static/queue/ns2.mp3")
					//Transcode next song
					go transcode(musicDir + "/" + queue[queuePos+2]["file"])
					cFile = 1
				}

				//updateQueue <- &updateQueueMsg{Song: song["Title"], Artist: song["Artist"]}
				//Let clients know that the current song is done and that we'll be pausing.
				//Also give them info about the next song to be played
				//During this time, clients that have not done so will transfer from livestream to downloads
				msg := map[string]string{"cmd": "done", "Title": song["Title"], "Artist": song["Artist"], "Album": song["Album"], "Cover": "/art/" + GetAlbumDir(song["file"]), "Time": song["Time"]}
				jsonMsg, _ := json.Marshal(msg)
				h.broadcast <- []byte(jsonMsg)

				conn.Pause(true)

				//Wait 3 seconds for clients to load the next song if necessary, then resume next song
				time.Sleep(1000 * time.Millisecond)

				if err != nil {
					fmt.Println("Couldn't get current song! Error: " + err.Error())
				} else {
					//Tell clients to begin the song
					msg = map[string]string{"cmd": "NS"}
					jsonMsg, _ := json.Marshal(msg)
					h.broadcast <- []byte(jsonMsg)
					conn.Pause(false)
				}
			}

		case msg := <-utaChan:
			if msg == "cfile" {
				utaChan <- strconv.Itoa(cFile)
			}
			if msg == "queue" {
				if len(queue[queuePos+1:]) > 0 {
					jsonMsg, err := json.Marshal(queue[queuePos+1:])
					if err != nil {
						fmt.Println("Warning, could not jsonify queue")
					}
					utaChan <- string(jsonMsg)
				} else {
					utaChan <- ""
				}
			}
		}
	}
}

func startUp() (library []mpd.Attrs, queue []mpd.Attrs) {
	var conn *mpd.Client

	fmt.Println("Connecting to MPD")
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	defer conn.Close()

	conn.Pause(true)

	status, _ := conn.Status()

	nsongpos64, _ := strconv.ParseInt(status["nextsong"], 10, 0)
	nsongpos := int(nsongpos64)

	pl, _ := conn.PlaylistInfo(-1, -1)
	psize := len(pl) - 1

	queue, err = conn.PlaylistInfo(nsongpos-1, psize)
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	nsong := queue[0]

	fmt.Println("Performing init transcodes")
	os.Remove("static/queue/next.mp3")
	transcode(musicDir + "/" + nsong["file"])
	os.Rename("static/queue/next.mp3", "static/queue/ns1.mp3")
	transcode(musicDir + "/" + queue[1]["file"])
	os.Rename("static/queue/next.mp3", "static/queue/ns2.mp3")
	transcode(musicDir + "/" + queue[2]["file"])

	conn.Pause(false)
	songs, err := conn.ListAllInfo("/")
	return songs, queue
}
