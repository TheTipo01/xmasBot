package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// cmdsStart starts all the exec.Cmd inside the slice
func cmdsStart(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		_ = cmd.Start()
	}
}

// cmdsWait waits for all the exec.Cmd inside the slice to finish processing, to free up resources
func cmdsWait(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		_ = cmd.Wait()
	}
}

// download downloads the song and gives back a pipe with DCA audio
func download(link string) []*exec.Cmd {
	// Starts yt-dlp with the arguments to select the best audio
	ytDlp := exec.Command("yt-dlp", "-q", "-f", "bestaudio*", "-a", "-", "-o", "-", "--geo-bypass",
		"--sponsorblock-remove", "sponsor,music_offtopic,outro")
	ytDlp.Stdin = strings.NewReader(link)
	ytOut, _ := ytDlp.StdoutPipe()

	// We pass it down to ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", "pipe:0", "-c:a", "libopus",
		"-b:a", "96k", "-frame_duration", strconv.Itoa(frameDuration), "-vbr", "off", "-f", "opus", "-")
	ffmpeg.Stdin = ytOut

	return []*exec.Cmd{ytDlp, ffmpeg}
}
