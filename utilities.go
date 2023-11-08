package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/lit"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func downloadSong(link string) error {
	mutex.Lock()
	defer mutex.Unlock()

	// Gets info about songs
	out, err := exec.Command("yt-dlp", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		return errors.New(fmt.Sprintf("Can't get info about song: %s", splittedOut[len(splittedOut)-1]))
	}

	// Check if youtube-dl returned something
	if strings.TrimSpace(splittedOut[0]) == "" {
		lit.Error("yt-dlp returned no songs")
		return errors.New("yt-dlp returned no songs")
	}

	var ytdl YoutubeDL

	// We parse every track as individual json, because youtube-dl
	for _, singleJSON := range splittedOut {
		_ = json.Unmarshal([]byte(singleJSON), &ytdl)

		cmds := download(ytdl.WebpageURL)
		fileName := ytdl.ID + "-" + ytdl.Extractor

		if ytdl.Extractor == "generic" {
			fileName = idGen(ytdl.WebpageURL) + "-" + ytdl.Extractor
		}

		// Opens the file, writes file to it, closes it
		file, _ := os.OpenFile(cachePath+fileName+audioExtension, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		cmds[len(cmds)-1].Stdout = file

		cmdsStart(cmds)
		cmdsWait(cmds)
		_ = file.Close()
		files = append(files, fileName+audioExtension)
	}

	return nil
}

// idGen returns the first 11 characters of the SHA1 hash for the given link
func idGen(link string) string {
	h := sha1.New()
	h.Write([]byte(link))

	return strings.ToLower(base32.HexEncoding.EncodeToString(h.Sum(nil))[0:11])
}

// Checks if a string is a valid URL
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}
