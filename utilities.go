package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/lit"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func downloadSong(link string) error {
	// Gets info about songs
	out, err := exec.Command("youtube-dl", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		err := fmt.Sprintf("Can't get info about song: %s", splittedOut[len(splittedOut)-1])

		lit.Error(err)
		return errors.New(err)
	}

	// Check if youtube-dl returned something
	if strings.TrimSpace(splittedOut[0]) == "" {
		lit.Error("youtube-dl returned no songs")
		return errors.New("youtube-dl returned no songs")
	}

	var ytdl YoutubeDL

	// We parse every track as individual json, because youtube-dl
	for _, singleJSON := range splittedOut {
		_ = json.Unmarshal([]byte(singleJSON), &ytdl)
		fileName := ytdl.ID + "-" + ytdl.Extractor

		// Checks if video is already downloaded
		info, err := os.Stat("./audio_cache/" + fileName + ".dca")

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			var cmd *exec.Cmd

			// Download and conversion to DCA
			switch runtime.GOOS {
			case "windows":
				cmd = exec.Command("gen.bat", fileName)
			default:
				cmd = exec.Command("sh", "gen.sh", fileName)
			}

			cmd.Stdin = strings.NewReader(ytdl.WebpageURL)
			out, err = cmd.CombinedOutput()

			if err != nil {
				splitted := strings.Split(string(out), "\n")

				err := fmt.Sprintf("Can't download song: %s", splitted[len(splitted)-1])

				lit.Error(err)
				return errors.New(err)
			}
		}

	}

	return nil
}
