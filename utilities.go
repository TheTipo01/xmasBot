package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func downloadSong(link string) error {
	// Gets info about songs
	out, err := exec.Command("yt-dlp", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		err := fmt.Sprintf("Can't get info about song: %s", splittedOut[len(splittedOut)-1])

		lit.Error(err)
		return errors.New(err)
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
		cmds[2].Stdout = file

		mutex.Lock()
		cmdsStart(cmds)
		cmdsWait(cmds)
		_ = file.Close()
		files = append(files, fileName+audioExtension)
		mutex.Unlock()
	}

	return nil
}

// idGen returns the first 11 characters of the SHA1 hash for the given link
func idGen(link string) string {
	h := sha1.New()
	h.Write([]byte(link))

	return strings.ToLower(base32.HexEncoding.EncodeToString(h.Sum(nil))[0:11])
}

// isCommandEqual compares two command by marshalling them to JSON. Yes, I know. I don't want to write recursive things.
func isCommandEqual(c *discordgo.ApplicationCommand, v *discordgo.ApplicationCommand) bool {
	c.Version = ""
	c.ID = ""
	c.ApplicationID = ""
	c.Type = 0
	cBytes, _ := json.Marshal(&c)

	v.Version = ""
	v.ID = ""
	v.ApplicationID = ""
	v.Type = 0
	vBytes, _ := json.Marshal(&v)

	return bytes.Compare(cBytes, vBytes) == 0
}

// Checks if a string is a valid URL
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

func sendEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c *chan int) {
	sliceEmbed := []*discordgo.MessageEmbed{embed}
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Embeds: sliceEmbed}})
	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		*c <- 1
	}
}

func sendAndDeleteEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	sendEmbedInteraction(s, embed, i, nil)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Modify an already sent interaction
func modifyInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction) {
	sliceEmbed := []*discordgo.MessageEmbed{embed}
	_, err := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{Embeds: &sliceEmbed})
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}
}

// Modify an already sent interaction and deletes it after the specified wait time
func modifyInteractionAndDelete(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	modifyInteraction(s, embed, i)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}
