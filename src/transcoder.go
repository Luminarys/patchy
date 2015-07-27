package main

import (
	"fmt"
	"os"
	"os/exec"
)

func transcode(song string) {
	cmd := "ffmpeg"
	args := []string{"-i", song, "-q:a", "2", "static/queue/next.mp3"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Fatal Error, could not transcode song! Additional info: "+err.Error())
		os.Exit(1)
	}
}
