package main

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/voice"
)

type Server struct {
	channel discord.ChannelID
	vs      *voice.Session
	errors  uint8
}

// YoutubeDL structure for holding youtube-dl data
type YoutubeDL struct {
	Extractor  string `json:"extractor"`
	ID         string `json:"id"`
	WebpageURL string `json:"webpage_url"`
}

type Config struct {
	Token   string `fig:"token" validate:"required"`
	Servers []struct {
		Guild   string `fig:"guild" validate:"required"`
		Channel string `fig:"channel" validate:"required"`
	}
	Admin []string `fig:"admin" validate:"required"`
}
