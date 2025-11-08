package main

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
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

// SendEmbedInteraction sends an embed as response to an interaction
func SendEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, c chan<- struct{}, isDeferred chan struct{}) {
	var err error

	if isDeferred != nil {
		<-isDeferred
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().SetEmbeds(embed).Build())
	} else {
		err = e.CreateMessage(discord.NewMessageCreateBuilder().SetEmbeds(embed).Build())
	}

	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		c <- struct{}{}
	}
}

// SendAndDeleteEmbedInteraction sends and deletes after three second an embed in a given channel
func SendAndDeleteEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration, isDeferred chan struct{}) {
	SendEmbedInteraction(embed, e, nil, isDeferred)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// ModifyInteraction modifies an already sent interaction
func ModifyInteraction(e *events.ApplicationCommandInteractionCreate, embed discord.Embed) {
	_, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().SetEmbeds(embed).Build())
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}
}

// ModifyInteractionAndDelete modifies an already sent interaction and deletes it after the specified wait time
func ModifyInteractionAndDelete(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration) {
	ModifyInteraction(e, embed)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}
