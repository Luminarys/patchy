package main

import (
	"fmt"
	"os"
	"os/exec"
)

func transcode(song string) {
	cmd := "ffmpeg"
	args := []string{"-i", song, "-threads", "0", "-b:a", "190k", "static/queue/next.mp3"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Fatal Error, could not transcode song! Additional info: "+err.Error())
		os.Exit(1)
	}
}
