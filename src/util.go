package main

import (
	"encoding/json"
	"github.com/fhs/gompd/mpd"
	"math/rand"
	"os"
	"strings"
	"time"
)

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

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
