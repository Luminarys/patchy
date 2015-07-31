package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type request struct {
	Title  string
	Album  string
	Artist string
}

func handleSongs(utaChan chan string, reChan chan string, l *library, h *hub, q *queue) {
	ctChan := make(chan int)
	q.playing = false
	lastTime := 0
	for msg := range utaChan {
		fmt.Println(msg)
		//Trigger this if there is a new song to be played
		if msg == "ns" && !q.playing {
			q.playing = true
			var ns *qsong
			//We'll want to make this concurrent since otherwise any requests which come in during the meanwhile will get pushed on wait
			//Needs some thought as to whether or not this is the best way to do stuff
			go func() {
				//Wait until all transcoding is done
				for q.transcoding {
				}
				//Precondition: q.queue has at least 1 item in it.
				//Consume item in queue, if there's anything left, initiate a transcode
				ns = q.consume()
				if len(q.queue) > 0 {
					fmt.Println("Queue has more than one item, performing next transcode in background")
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
			}()
		}

		//Get current song file in use
		if msg == "cfile" {
			reChan <- strconv.Itoa(q.CFile)
		}

		//If a song just finished, load in the next thing from queue if available
		if msg == "done" {
			q.playing = false
			if len(q.queue) > 0 {
				go func() {
					utaChan <- "ns"
				}()
			}
		}

		if msg == "ctime" {
			if q.playing {
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
	}
}

func handleRequests(requests chan *request, utaChan chan string, q *queue, l *library, h *hub) {
	for req := range requests {
		song, err := l.reqSearch(req.Title, req.Album, req.Artist)
		if err != nil {
			fmt.Println("Couldn't add request error: " + err.Error())
		} else {
			q.add(song)
			msg := map[string]string{"cmd": "queue", "Title": song.Title, "Artist": song.Artist}
			jsonMsg, _ := json.Marshal(msg)
			h.broadcast <- []byte(jsonMsg)

			if len(q.queue) == 1 {
				if !q.playing {
					fmt.Println("Queue has only one item, performing transcode and sending ns")
					q.transcodeNext()
					utaChan <- "ns"
				} else {
					fmt.Println("Queue has only one item, performing transcode")
					q.transcodeNext()
				}
			}
		}
	}
}
