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
	"strconv"
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

func getSearchRes(ctx *web.Context, req string, l *library) string {
	res := l.asyncSearch(req)
	jsonMsg, _ := json.Marshal(res)
	return string(jsonMsg)
}

func getNowPlaying(ctx *web.Context, utaChan chan string, reChan chan string, queue *queue) string {
	song := make(map[string]string)

	if np := queue.np; np != nil {
		utaChan <- "ctime"
		ctime := <-reChan

		utaChan <- "cfile"
		cfile := <-reChan

		song["Title"] = np.Title
		song["Artist"] = np.Artist
		song["Album"] = np.Album
		song["file"] = np.File
		song["Time"] = strconv.Itoa(np.Length)

		song["ctime"] = ctime
		song["cfile"] = cfile
	} else {
		song["Title"] = "N/A"
		song["Artist"] = "N/A"
		song["Album"] = "N/A"
		song["file"] = "lol"
		song["Time"] = "0"

		song["ctime"] = "0"
		song["cfile"] = "1"
	}
	jsonMsg, _ := json.Marshal(song)
	return string(jsonMsg)
}

func getLibrary(ctx *web.Context, subset []mpd.Attrs) string {
	jsonMsg, _ := json.Marshal(subset)
	return string(jsonMsg)
}

func getQueue(ctx *web.Context, q *queue) string {
	//Let the song handler return a JSONify'd queue
	jsonMsg, _ := json.Marshal(q.queue)
	return string(jsonMsg)
}
