package manager

import (
	"encoding/binary"
	"errors"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
	"io"
	"os"
)

// Plays a song in DCA format
func playSound(el *queue.Element, server *Server) (SkipReason, error) {
	var (
		opuslen    int16
		skip       bool
		skipReason SkipReason
		err        error
	)

	// Start speaking.
	_ = server.VC.Speaking(true)

	for {
		select {
		case <-server.Pause:
			select {
			case <-server.Resume:
				el.Segments = server.Queue.GetFirstElement().Segments
			case skipReason = <-server.Skip:
				cleanUp(server, el.Closer)
				return skipReason, nil
			}
		case skipReason = <-server.Skip:
			cleanUp(server, el.Closer)
			return skipReason, nil
		default:
			if el.Segments[server.Frames] {
				skip = !skip
			}

			// Read opus frame length from dca file.
			err = binary.Read(el.Reader, binary.LittleEndian, &opuslen)

			// If this is the end of the file, just return.
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				if el.Loop {
					if el.Closer != nil {
						_ = el.Closer.Close()
					}

					f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
					el.Reader = f
					el.Closer = f
					continue
				} else {
					cleanUp(server, el.Closer)
					return Finished, nil
				}
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opuslen)
			err = binary.Read(el.Reader, binary.LittleEndian, &InBuf)

			// Keep count of the frames in the song
			el.Frames++

			if skip {
				continue
			}

			// Should not be any end of file errors
			if err != nil {
				cleanUp(server, el.Closer)
				// Force to re-download the song
				_ = os.Remove(constants.CachePath + el.ID + ".dca")
				return Error, err
			}

			server.VC.OpusSend <- InBuf
		}
	}
}

func cleanUp(server *Server, closer io.Closer) {
	_ = server.VC.Speaking(false)
	server.Frames = 0

	if closer != nil {
		_ = closer.Close()
	}
}
