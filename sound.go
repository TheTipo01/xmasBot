package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/bwmarrin/lit"
	"io"
	"os"
	"sync"
)

func playSound(fileName string) {
	var opuslen int16

	file, err := os.Open(cachePath + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}
	defer file.Close()

	buffer := bufio.NewReader(file)
	wg := sync.WaitGroup{}

	for {
		// Read opus frame length from dca file.
		err = binary.Read(buffer, binary.LittleEndian, &opuslen)

		// If this is the end of the file, return.
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(buffer, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		wg.Wait()
		wg.Add(len(servers))
		for _, i := range servers {
			go func(i *Server) {
				defer wg.Done()
				i.vc.OpusSend <- InBuf
			}(i)
		}
	}
}
