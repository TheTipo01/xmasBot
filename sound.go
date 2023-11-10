package main

import (
	"github.com/bwmarrin/lit"
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

	if err = oggreader.DecodeBuffered(MiddlemanWriter{}, file); err != nil {
		lit.Error("Error playing opus file: %s", err)
	}
}

type MiddlemanWriter struct{}

func (m MiddlemanWriter) Write(p []byte) (n int, err error) {
	serversMutex.Lock()
	defer serversMutex.Unlock()

	// Try to write what we received to the master writer
	for g, s := range servers {
		_, err := s.vs.Write(p)
		if err != nil {
			s.errors++
			if s.errors >= 3 {
				// Try to reconnect
				reconnect(g)
			}
		} else {
			s.errors = 0
		}
	}

	return len(p), nil
}
