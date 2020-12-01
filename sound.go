package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os"
)

func playSound(vc *discordgo.VoiceConnection, fileName string) {
	var opuslen int16

	file, err := os.Open("./audio_cache/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		return
	}

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

		// Stream data to discord
		vc.OpusSend <- InBuf

	}

	// Close the file
	_ = file.Close()

}
