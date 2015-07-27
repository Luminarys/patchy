package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"github.com/hoisie/web"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var musicDir string = "/home/eumen/Music"
var queuePos int
var cFile int = 1

func main() {
	var conn *mpd.Client
	fmt.Println("Connecting to MPD")
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

	//Get init queue and do the initial transcode
	status, err := conn.Status()
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	conn.Pause(true)
	nsongpos64, _ := strconv.ParseInt(status["nextsong"], 10, 0)
	nsongpos := int(nsongpos64)

	pl, _ := conn.PlaylistInfo(-1, -1)
	psize := len(pl) - 1

	queue, err := conn.PlaylistInfo(nsongpos-1, psize)
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	queuePos = 0
	nsong := queue[queuePos]

	fmt.Println("Performing init transcodes")
	os.Remove("static/queue/next.mp3")
	transcode(musicDir + "/" + nsong["file"])
	os.Rename("static/queue/next.mp3", "static/queue/ns1.mp3")
	transcode(musicDir + "/" + queue[1]["file"])
	os.Rename("static/queue/next.mp3", "static/queue/ns2.mp3")
	transcode(musicDir + "/" + queue[2]["file"])

	conn.Pause(false)

	h := newHub()
	go h.run()

	// Log errors.
	go func() {
		for err := range w.Error {
			log.Println("Error:", err)
		}
	}()

	//Control song transitions -- During this time, update the websockets
	go func() {
		var status mpd.Attrs
		for _ = range w.Event {
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
			if pos == 0.000 {
				//Stop us from getting into an infinite loop by waiting 25 ms
				time.Sleep(25 * time.Millisecond)
				song, err := conn.CurrentSong()
				//Prep next song
				queuePos++
				fmt.Println("Next song:")
				fmt.Println(queue[queuePos])
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

				//Wait 5 seconds for clients to load the next song if necessary, then resume next song
				time.Sleep(5000 * time.Millisecond)

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
		}
	}()

	songs, err := conn.ListAllInfo("/")
	shuffle(songs)
	subset := songs[:20]

	//Searches for cover image
	web.Get("/art/(.+)", getCover)

	//Gets the song -- apparently firefox is a PoS and needs manual header setting
	web.Get("/queue/(.+)", func(ctx *web.Context, song string) string {
		//Open the file
		f, err := os.Open("static/queue/" + song)
		if err != nil {
			return "Error reading file!\n"
		}

		//Get MIME
		r, err := ioutil.ReadAll(f)
		if err != nil {
			return "Error reading file!\n"
		}
		mime := http.DetectContentType(r)

		_, err = f.Seek(0, 0)
		if err != nil {
			return "Error reading the file\n"
		}
		//This is weird - ServeContent supposedly handles MIME setting
		//But the Webgo content setter needs to be used too
		//In addition, ServeFile doesn't work, ServeContent has to be used
		ctx.ContentType(mime)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "static/queue/"+song, time.Now(), f)
		return ""
	})

	//Returns main page with custom selection of songs
	web.Get("/", func(ctx *web.Context) string {
		return getIndex(ctx, subset)
	})

	//Returns a raw song
	web.Get("/song/(.+)", getSong)

	//Returns the JSON info for the currently playing song
	web.Get("/np", func(ctx *web.Context) string {
		song, err := conn.CurrentSong()
		if err != nil {
			fmt.Println("Couldn't get current status! Error: " + err.Error())
			conn.Close()

			fmt.Println("Reconnecting...")
			conn, err = mpd.Dial("tcp", "127.0.0.1:6600")
			if err != nil {
				fmt.Println("Error: could not connect to MPD, exiting")
				os.Exit(1)
			}
			song, _ = conn.CurrentSong()
		}
		status, _ := conn.Status()
		fmt.Println(status)
		ctime := strings.SplitAfterN(status["time"], ":", 2)[0]
		last := len(ctime) - 1
		song["ctime"] = ctime[:last]
		song["cfile"] = strconv.Itoa(cFile)
		jsonMsg, _ := json.Marshal(song)
		return string(jsonMsg)
	})

	//Handle the websocket
	web.Websocket("/ws", websocket.Handler(func(ws *websocket.Conn) {
		handleSocket(ws, h)
	}))
	web.Get("/library", func(ctx *web.Context) string {
		jsonMsg, _ := json.Marshal(subset)
		return string(jsonMsg)
	})
	web.Run("0.0.0.0:8080")
}

func getIndex(ctx *web.Context, songs []mpd.Attrs) string {
	funcMap := template.FuncMap{
		"AlbumDir": GetAlbumDir,
	}
	t, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Couldn't parse template! Error: " + err.Error())
	}
	t = t.Funcs(funcMap)
	if err != nil {
		fmt.Println("Couldn't parse template! Error: " + err.Error())
	}
	if err = t.Execute(ctx.ResponseWriter, songs); err != nil {
		fmt.Println("Couldn't execute template! Error: " + err.Error())
	}
	return ""
}

func getSong(ctx *web.Context, songLoc string) string {
	song := musicDir + "/" + songLoc
	f, err := os.Open(song)
	if err != nil {
		return "Error reading file!\n"
	}

	//Get MIME
	r, err := ioutil.ReadAll(f)
	if err != nil {
		return "Error reading file!\n"
	}
	mime := http.DetectContentType(r)

	_, err = f.Seek(0, 0)
	if err != nil {
		return "Error reading the file\n"
	}
	ctx.ContentType(mime)
	http.ServeContent(ctx.ResponseWriter, ctx.Request, song, time.Now(), f)
	return ""
}

func getCover(ctx *web.Context, album string) string {
	dir := musicDir + "/" + album
	cover := "static/image/missing.png"
	//Do various searches -- Optimally this should do a full traversal and find one of these names
	if exists(dir + "/cover.jpg") {
		cover = dir + "/cover.jpg"
	} else if exists(dir + "/cover.png") {
		cover = dir + "/cover.png"
	} else if exists(dir + "/folder.png") {
		cover = dir + "/folder.png"
	} else if exists(dir + "/folder.jpg") {
		cover = dir + "/folder.jpg"
	}
	//Open the file
	f, err := os.Open(cover)
	if err != nil {
		return "Error reading file!\n"
	}

	//Get MIME
	r, err := ioutil.ReadAll(f)
	if err != nil {
		return "Error reading file!\n"
	}
	mime := http.DetectContentType(r)

	_, err = f.Seek(0, 0)
	if err != nil {
		return "Error reading the file\n"
	}
	//This is weird - ServeContent supposedly handles MIME setting
	//But the Webgo content setter needs to be used too
	//In addition, ServeFile doesn't work, ServeContent has to be used
	ctx.ContentType(mime)
	http.ServeContent(ctx.ResponseWriter, ctx.Request, cover, time.Now(), f)
	return ""
}

func shuffle(arr []mpd.Attrs) {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond())) // no shuffling without this line

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func GetAlbumDir(song string) string {
	return strings.SplitAfterN(song, "/", 2)[0]
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
