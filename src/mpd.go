package main

import (
	"encoding/json"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"os"
	"strconv"
	"time"
)

func handleSongs(utaChan chan string, reChan chan string, library []mpd.Attrs, h *hub, q *queue) {
	ctChan := make(chan int)
	started := false
	lastTime := 0
	for msg := range utaChan {
		fmt.Println(msg)
		//Trigger this if there is a new song to be played
		if msg == "ns" {
			started = true
			var ns *qsong
			//If there's only one thing in the queue, transcode it and then consume it
			//Precondition: q.queue has at least 1 item in it.
			ns = q.consume()
			if len(q.queue) > 1 {
				go q.transcodeNext()
			}
			//updateQueue <- &updateQueueMsg{Song: song["Title"], Artist: song["Artist"]}
			//Let clients know that the current song is done and that we'll be pausing.
			//Also give them info about the next song to be played
			//During this time, clients that have not done so will transfer from livestream to downloads
			msg := map[string]string{"cmd": "done"}
			lastTime = ns.Length
			jsonMsg, _ := json.Marshal(msg)
			h.broadcast <- []byte(jsonMsg)

			//Wait 2 seconds for clients to load the next song if necessary, then resume next song
			time.Sleep(2000 * time.Millisecond)

			//Tell clients to begin the song
			msg = map[string]string{"cmd": "NS", "Title": ns.Title, "Artist": ns.Artist, "Album": ns.Album, "Cover": "/art/" + GetAlbumDir(ns.File), "Time": strconv.Itoa(ns.Length)}
			jsonMsg, _ = json.Marshal(msg)
			h.broadcast <- []byte(jsonMsg)
			go timer(ns.Length, utaChan, ctChan)
		}

		//Get current song file in use
		if msg == "cfile" {
			reChan <- strconv.Itoa(q.CFile)
		}

		//If a song just finished, load in the next thing from queue if available
		if msg == "done" {
			started = false
			if len(q.queue) > 0 {
				go func() {
					utaChan <- "ns"
				}()
			}
		}

		if msg == "ctime" {
			if started {
				ctChan <- 0
				reChan <- strconv.Itoa(<-ctChan)
			} else {
				//We want to actually do 100% here, do it later >.>
				reChan <- strconv.Itoa(lastTime)
			}
		}

		/*
			if msg == "queue" {
				jsonMsg, err := json.Marshal(q.queue)
				if err != nil {
					fmt.Println("Warning, could not jsonify queue")
				}
				utaChan <- string(jsonMsg)
			} else {
				utaChan <- ""
			}
		*/

		//Handles requests
		if isJSON(msg) {
			var req map[string]string
			if err := json.Unmarshal([]byte(msg), &req); err != nil {
				fmt.Println("Error, couldn't unmarshal client request")
			} else {
				search(req, h, utaChan, library, q, started)
			}
		}
	}
}

func search(req map[string]string, h *hub, utaChan chan string, songs []mpd.Attrs, q *queue, playing bool) {
	for _, song := range songs {
		if song["Title"] == req["Title"] && (song["Album"] == req["Album"] || song["Artist"] == req["Artist"]) {
			fmt.Println("Found song: " + song["file"])
			st, err := strconv.Atoi(song["Time"])
			if err != nil {
				fmt.Println("Couldn't add song due to time conversion error!")
				break
			}
			q.add(&qsong{Title: song["Title"], Album: song["Album"], Artist: song["Artist"], Length: st, File: song["file"]})

			msg := map[string]string{"cmd": "queue", "Title": song["Title"], "Artist": song["Artist"]}
			jsonMsg, _ := json.Marshal(msg)
			h.broadcast <- []byte(jsonMsg)

			if len(q.queue) == 1 {
				if !playing {
					go func() {
						q.transcodeNext()
						utaChan <- "ns"
					}()
				} else {
					go q.transcodeNext()
				}
			}
			break
		}
	}
}

func startUp() (library []mpd.Attrs) {
	var conn *mpd.Client

	fmt.Println("Connecting to MPD")
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	defer conn.Close()

	songs, err := conn.ListAllInfo("/")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	os.Remove("static/queue/ns1.mp3")
	os.Remove("static/queue/ns2.mp3")
	os.Remove("static/queue/next.mp3")
	return songs
}
