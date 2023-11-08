package main

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/voice"
)

type Server struct {
	channel discord.ChannelID
	vs      *voice.Session
}

// YoutubeDL structure for holding youtube-dl data
type YoutubeDL struct {
	Extractor  string `json:"extractor"`
	ID         string `json:"id"`
	WebpageURL string `json:"webpage_url"`
}
