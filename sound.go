package main

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/lit"
)

func playSound(fileName string) {
	file, err := os.Open(cachePath + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}
	defer file.Close()

	wg := sync.WaitGroup{}

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	var frameLen int16
	// Don't wait for the first tick, run immediately.
	for ; true; <-ticker.C {
		err = binary.Read(file, binary.LittleEndian, &frameLen)
		if err != nil {
			if err == io.EOF {
				_ = file.Close()
				return
			}
			panic("error reading file: " + err.Error())
			return
		}

		wg.Wait()
		for _, i := range servers {
			wg.Add(1)
			go func(i *Server) {
				defer wg.Done()

				// Copy the frame.
				_, err = io.CopyN(i.vc.UDP(), file, int64(frameLen))
				if err != nil && err != io.EOF {
					_ = file.Close()
					return
				}
			}(i)
		}

	}
}
