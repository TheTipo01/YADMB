package main

import (
	"encoding/binary"
	"github.com/bwmarrin/lit"
	"io"
)

// Plays a song from a io.Reader if specified, or tries to open a file with the given fileName
func playSound(guildID string, in io.Reader) {
	var (
		opuslen int16
		skip    bool
		err     error
	)

	server[guildID].skip = false

	// Start speaking.
	_ = server[guildID].vc.Speaking(true)

	for {
		if server[guildID].queue.GetFirstElement().Segments[server[guildID].frames] {
			skip = !skip
		}

		// Read opus frame length from dca file.
		err = binary.Read(in, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			lit.Debug(err.Error())
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(in, binary.LittleEndian, &InBuf)

		// Keep count of the frames of the song
		server[guildID].frames++

		if skip {
			continue
		}

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error streaming from dca file: %s", err)
			break
		}

		if !server[guildID].skip {
			server[guildID].vc.OpusSend <- InBuf
		} else {
			break
		}
	}

	_ = server[guildID].vc.Speaking(false)
	server[guildID].frames = 0
	lit.Debug("Finished playing song %s", server[guildID].queue.GetFirstElement().ID)
}
