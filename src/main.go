package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hoisie/web"
	"os"
)

var settings struct {
	MusicDir string `json:"music_dir"`
	Port     string `json:"port"`
}

var musicDir string

func main() {

	configFile, err := os.Open("patchy.conf")
	if err != nil {
		fmt.Println("Couldn't open conf file!")
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&settings); err != nil {
		fmt.Println("Couldn't parse conf file!", err.Error())
		os.Exit(1)
	}

	port := flag.String("port", settings.Port, "The port that patchy listens on.")
	musicDirFlag := flag.String("mdir", settings.MusicDir, "The full filepath to the mpd library location.")
	flag.Parse()

	musicDir = *musicDirFlag

	startUp()

	h := newHub()
	go h.run()

	q := newQueue()

	l := newLibrary()
	subset := l.selection()

	//Control song transitions -- During this time, update the websockets and notify clients
	utaChan := make(chan string)
	reChan := make(chan string)
	go handleSongs(utaChan, reChan, l, h, q)

	requests := make(chan *request)
	go handleRequests(requests, utaChan, q, l, h)

	//Searches for cover image
	web.Get("/art/(.+)", getCover)

	//Gets the song -- apparently firefox is a PoS and needs manual header setting
	web.Get("/queue/(.+)", getSong)

	//Search for songs similar to a given title
	web.Get("/search/(.+)", func(ctx *web.Context, req string) string {
		return getSearchRes(ctx, req, l)
	})

	//Returns the JSON info for the currently playing song
	web.Get("/np", func(ctx *web.Context) string {
		return getNowPlaying(ctx, utaChan, reChan, q, len(h.connections))
	})

	//Handle the websocket
	web.Websocket("/ws", websocket.Handler(func(ws *websocket.Conn) {
		handleSocket(ws, h, utaChan, requests)
	}))

	//Returns a library sample for initial client display
	web.Get("/library", func(ctx *web.Context) string {
		return getLibrary(ctx, subset)
	})

	//Returns the current queue
	web.Get("/curQueue", func(ctx *web.Context) string {
		return getQueue(ctx, q)
	})

	//Returns the current queue
	web.Post("/upload", func(ctx *web.Context) string {
		return handleUpload(ctx, l)
	})

	web.Run("0.0.0.0:" + *port)
}

func startUp() {
	os.Remove("static/queue/ns1.mp3")
	os.Remove("static/queue/ns2.mp3")
	os.Remove("static/queue/next.mp3")
}
