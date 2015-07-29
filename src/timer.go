package main

import (
	"time"
)

//Timer which sends true into a done channel upon completion
//And can be queried for the current duration at any time until completion
func timer(duration int, done chan string, current chan int) {
	startTime := time.Now()
	over := make(chan bool)
	go wait(duration, over)
	for {
		select {
		case <-over:
			done <- "done"
		case <-current:
			current <- int(time.Since(startTime).Seconds())
		}
	}
}

func wait(duration int, over chan bool) {
	time.Sleep(time.Duration(duration) * time.Second)
	over <- true
}
