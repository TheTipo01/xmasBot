package main

import (
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

type Server struct {
	channel snowflake.ID
	vc      voice.Conn
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
	Admin  []string `fig:"admin" validate:"required"`
	Status string   `fig:"status" validate:"required"`
}
