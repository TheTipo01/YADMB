package main

import (
	"encoding/binary"
	"github.com/TheTipo01/YADMB/Queue"
	"io"
	"os"
)

// Plays a song in DCA format
func playSound(guildID string, el *Queue.Element) bool {
	var (
		opuslen int16
		skip    bool
		err     error
	)

	// Start speaking.
	_ = server[guildID].vc.Speaking(true)

	for {
		select {
		case <-server[guildID].pause:
			select {
			case <-server[guildID].resume:
				el.Segments = server[guildID].queue.GetFirstElement().Segments
			case <-server[guildID].skip:
				cleanUp(guildID, el.Closer)
				return true
			}
		case <-server[guildID].skip:
			cleanUp(guildID, el.Closer)
			return true
		default:
			if el.Segments[server[guildID].frames] {
				skip = !skip
			}

			// Read opus frame length from dca file.
			err = binary.Read(el.Reader, binary.LittleEndian, &opuslen)

			// If this is the end of the file, just return.
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				if el.Loop {
					if el.Closer != nil {
						_ = el.Closer.Close()
					}

					f, _ := os.Open(cachePath + el.ID + ".dca")
					el.Reader = f
					el.Closer = f
					continue
				} else {
					cleanUp(guildID, el.Closer)
					return false
				}
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opuslen)
			err = binary.Read(el.Reader, binary.LittleEndian, &InBuf)

			// Keep count of the frames in the song
			server[guildID].frames++

			if skip {
				continue
			}

			// Should not be any end of file errors
			if err != nil {
				cleanUp(guildID, el.Closer)
				// Force to re-download the song
				_ = os.Remove(cachePath + el.ID + ".dca")
				return false
			}

			server[guildID].vc.OpusSend <- InBuf
		}
	}
}

func cleanUp(guildID string, closer io.Closer) {
	_ = server[guildID].vc.Speaking(false)
	server[guildID].frames = 0

	if closer != nil {
		_ = closer.Close()
	}
}
