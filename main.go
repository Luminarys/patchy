package main

import (
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

var musicDir string = "/home/eumen/Music"

func main() {
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD, exiting")
		os.Exit(1)
	}
	song, err := conn.CurrentSong()
	if err != nil {
		fmt.Println("No Song!")
	} else {
		fmt.Println(song["Title"])
	}
	//Searches for cover image
	web.Get("/art/(.+)", getCover)
	web.Get("/", func(ctx *web.Context) string {
		return getIndex(ctx, conn)
	})
	web.Run("0.0.0.0:8080")
}

func getIndex(ctx *web.Context, conn *mpd.Client) string {
	funcMap := template.FuncMap{
		"AlbumDir": GetAlbumDir,
	}
	t, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Couldn't parse template! Error: " + err.Error())
	}
	t = t.Funcs(funcMap)
	songs, err := conn.ListAllInfo("/")
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
	if exists(dir + "/cover.jpg") {
		cover = dir + "/cover.jpg"
	} else if exists(dir + "/cover.png") {
		cover = dir + "/cover.png"
	} /* else {
		//Search one subdir deep for stuff
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Println(err)
		}
		for _, file := range files {
			f, err := os.Open(dir + "/" + file.Name())
			if err != nil {
				fmt.Println(err)
				//break
			}
			fi, err := f.Stat()
			if err != nil {
				fmt.Println(err)
				//break
			}
			if fi.Mode().IsDir() {
				dir = dir + "/" + file.Name()
				if exists(dir + "/cover.jpg") {
					cover = dir + "/cover.jpg"
				} else if exists(dir + "/cover.png") {
					cover = dir + "/cover.png"
				}
				break
			}
		}
	}
	*/
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

func GetAlbumDir(song string) string {
	fmt.Println(song)
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
