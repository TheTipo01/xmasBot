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
	for _, g := range servers {
		_, _ = g.vs.Write(p)
	}

	return len(p), nil
}
