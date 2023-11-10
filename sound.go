package main

import (
	"github.com/bwmarrin/lit"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/oggreader"
	"os"
)

func playSound(fileName string) {
	file, err := os.Open(cachePath + fileName)
	if err != nil {
		lit.Error("Error opening opus file: %s", err)
		return
	}
	defer file.Close()

	if err = oggreader.DecodeBuffered(NewMiddlemanWriter(), file); err != nil {
		lit.Error("Error playing opus file: %s", err)
	}
}

type MiddlemanWriter struct {
	errors map[discord.GuildID]uint32
}

func NewMiddlemanWriter() MiddlemanWriter {
	return MiddlemanWriter{
		errors: make(map[discord.GuildID]uint32),
	}
}

func (m MiddlemanWriter) Write(p []byte) (int, error) {
	serversMutex.Lock()
	defer serversMutex.Unlock()

	// Try to write what we received to the master writer
	for g, s := range servers {
		_, err := s.vs.Write(p)
		if err != nil {
			m.errors[g]++
			if m.errors[g] > 5 {
				// Try to reconnect
				reconnect(g)
			}
		} else {
			m.errors[g] = 0
		}
	}

	return len(p), nil
}
