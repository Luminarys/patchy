package main

import (
	"errors"
	"fmt"
	"github.com/fhs/gompd/mpd"
	"os"
	"strconv"
	"strings"
	"time"
)

type library struct {
	// The library.
	library []mpd.Attrs
}

//Create a new queue
func newLibrary() *library {
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
	shuffle(songs)

	return &library{
		library: songs,
	}
}

//returns a small selection of songs for initial display
func (l *library) selection() []mpd.Attrs {
	return l.library[:20]
}

//Searches for a request and returns the first song which matches
func (l *library) reqSearch(title string, album string, artist string) (*qsong, error) {
	for _, song := range l.library {
		if song["Title"] == title && (song["Album"] == album || song["Artist"] == artist) {
			fmt.Println("Found song: " + song["file"])
			st, err := strconv.Atoi(song["Time"])
			if err != nil {
				fmt.Println("Couldn't get song due to time conversion error!")
				return nil, errors.New("Couldn't convert Time to int!")
			}
			return &qsong{Title: song["Title"], Album: song["Album"], Artist: song["Artist"], Length: st, File: song["file"]}, nil
		}
	}
	return nil, errors.New("No songs found!")
}

func (l *library) asyncSearch(req string) []mpd.Attrs {
	res := make([]mpd.Attrs, 0)

	//There has to be a faster way to do this >.>
	for _, song := range l.library {
		if strings.Contains(song["Title"], req) || strings.Contains(song["Album"], req) || strings.Contains(song["Artist"], req) {
			res = append(res, song)
			if len(res) == 20 {
				break
			}
		}
	}
	return res
}

//Updates the library
func (l *library) update() error {
	var conn *mpd.Client

	fmt.Println("Connecting to MPD")
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		fmt.Println("Error: could not connect to MPD for lib update")
		return errors.New("Could not connect to MPD!")
	}
	defer conn.Close()

	_, err = conn.Update("")
	if err != nil {
		fmt.Println("Error: could not update library!")
		return err
	}

	//Let the update happen
	time.Sleep(2 * time.Second)
	songs, err := conn.ListAllInfo("/")
	if err != nil {
		fmt.Println("Error: could not retrieve new library!")
		return err
	}

	l.library = songs
	return nil
}
