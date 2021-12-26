package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os"
	"time"
)

func playSound(fileName string, s *discordgo.Session) {
	var opuslen int16

	file, err := os.Open(cachePath + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}

	// Channel to send ok messages
	c1 := make(chan string, 1)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		for guild, server := range servers {
			// Send data in a goroutine
			go func() {
				server.vc.OpusSend <- InBuf
				c1 <- "ok"
			}()

			// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
			select {
			case _ = <-c1:
				break
			case <-time.After(time.Second / 2):
				server.vc, _ = s.ChannelVoiceJoin(guild, server.channel, false, true)
			}
		}
	}

	// Close the file
	_ = file.Close()
}
