package main

import (
	"encoding/json"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"github.com/hoisie/web"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func getSong(ctx *web.Context, song string) string {
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
}

func getNowPlaying(ctx *web.Context, utaChan chan string) string {
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	song, _ := conn.CurrentSong()
	status, _ := conn.Status()
	fmt.Println(status)
	ctime := strings.SplitAfterN(status["time"], ":", 2)[0]
	last := len(ctime) - 1
	song["ctime"] = ctime[:last]
	utaChan <- "cfile"
	song["cfile"] = <-utaChan
	jsonMsg, _ := json.Marshal(song)
	return string(jsonMsg)
}

func getLibrary(ctx *web.Context, subset []mpd.Attrs) string {
	jsonMsg, _ := json.Marshal(subset)
	return string(jsonMsg)
}

func getQueue(ctx *web.Context, utaChan chan string) string {
	//Let the song handler return a JSONify'd queue
	utaChan <- "queue"
	return <-utaChan
}
