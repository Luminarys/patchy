package main

import (
	"fmt"
	"os"
)

type qsong struct {
	Title  string
	Album  string
	Artist string
	Length int
	File   string
}

type queue struct {
	// Queue.
	queue []*qsong
	//Current playing song
	np *qsong
	//Current file in use
	CFile int
	//Transcoding status
	transcoding bool
	//Playing status
	playing bool
	//Previously transcoded file -- Used to prevent dupes
	pt string
}

//Create a new queue
func newQueue() *queue {
	return &queue{
		queue:       make([]*qsong, 0),
		CFile:       1,
		np:          nil,
		transcoding: false,
		pt:          "",
	}
}

//Consumes and returns the first value in the queue
//Precondition: queue has at least one item in it
func (q *queue) consume() *qsong {
	/*
		if len(q.Queue) < 1 {
			return nil, errors.New("Nothing in queue!")
		}
	*/
	s := q.queue[0]
	if len(q.queue) > 1 {
		q.queue = q.queue[1:]
	} else {
		//There has to be a better way of doing this
		q.queue = make([]*qsong, 0)
	}
	if q.CFile == 1 {
		q.CFile = 2
	} else {
		q.CFile = 1
	}
	q.np = s
	return s
}

//Adds a new item to the queue
func (q *queue) add(s *qsong) {
	q.queue = append(q.queue, s)
}

//Transcodes the next appropriate song
func (q *queue) transcodeNext() {
	//Need a better way of doing this -- perhaps transfer
	//From a nontranscoded queue to a transcoded queue?
	if q.pt == q.queue[0].File {
		fmt.Println("This song has already been transcoded!")
		return
	}
	q.pt = q.queue[0].File

	fmt.Println("Transcoding Song: ", q.queue[0].File)
	q.transcoding = true
	transcode(musicDir + "/" + q.queue[0].File)
	//Rename to opposite of current file, since the clients will be told to go
	//to the next song after this
	//Transcodes will happen BEFORE consumes(need to create the file for client use)
	//if there is only one thing in the queue
	//other wise transcodes occur afterwards(since you want to transcodein background)
	if q.CFile == 1 {
		fmt.Println("Renaming Song to ns2.mp3")
		os.Rename("static/queue/next.mp3", "static/queue/ns2.mp3")
	} else {
		fmt.Println("Renaming Song to ns1.mp3")
		os.Rename("static/queue/next.mp3", "static/queue/ns1.mp3")
	}
	q.transcoding = false
}
