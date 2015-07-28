package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/hoisie/web"
)

const musicDir string = "/home/eumen/Music"

func main() {
	songs, queue := startUp()
	shuffle(songs)
	subset := songs[:20]

	h := newHub()
	go h.run()

	//Control song transitions -- During this time, update the websockets and notify clients
	utaChan := make(chan string)
	go handleSongs(utaChan, queue, h)

	//Searches for cover image
	web.Get("/art/(.+)", getCover)

	//Gets the song -- apparently firefox is a PoS and needs manual header setting
	web.Get("/queue/(.+)", getSong)

	//Returns main page with custom selection of songs
	web.Get("/", func(ctx *web.Context) string {
		return getIndex(ctx, subset)
	})

	//Returns the JSON info for the currently playing song
	web.Get("/np", func(ctx *web.Context) string {
		return getNowPlaying(ctx, utaChan)
	})

	//Handle the websocket
	web.Websocket("/ws", websocket.Handler(func(ws *websocket.Conn) {
		handleSocket(ws, h, utaChan)
	}))

	//Returns a library sample for initial client display
	web.Get("/library", func(ctx *web.Context) string {
		return getLibrary(ctx, subset)
	})

	//Returns the current queue
	web.Get("/curQueue", func(ctx *web.Context) string {
		return getQueue(ctx, utaChan)
	})

	web.Run("0.0.0.0:8080")
}
