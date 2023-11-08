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
	writerMutex.Lock()

	// Try to write what we received to the master writer
	n, err = w.Write(p)
	if err != nil {
		writerMutex.Unlock()

		// If we can't write, we're probably disconnected or moved: wait for the voiceStateUpdate to recreate the writer
		<-done
		return m.Write(p)
	} else {
		writerMutex.Unlock()

		return n, err
	}
}
