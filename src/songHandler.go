package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func handleSongs(utaChan chan string, reChan chan string, l *library, h *hub, q *queue) {
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
			lastTime = ns.Length

			msg := map[string]string{"cmd": "done"}
			jsonMsg, _ := json.Marshal(msg)
			h.broadcast <- []byte(jsonMsg)

			//Wait 4 seconds for clients to load the next song if necessary, then resume next song
			time.Sleep(4000 * time.Millisecond)

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
				if started {
					ctChan <- 0
					ctime := <-ctChan
					if len(q.queue) != 0 || q.np.Length-ctime > 15 {
						search(req, h, utaChan, l, q, started)
					}
				} else {
					search(req, h, utaChan, l, q, started)
				}
			}
		}
	}
}

func search(req map[string]string, h *hub, utaChan chan string, l *library, q *queue, playing bool) {
	song, err := l.reqSearch(req["Title"], req["Album"], req["Artist"])
	if err != nil {
		fmt.Println("Couldn't add request error: " + err.Error())
	} else {
		q.add(song)
		msg := map[string]string{"cmd": "queue", "Title": song.Title, "Artist": song.Artist}
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
	}
}
