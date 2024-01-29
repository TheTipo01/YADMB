package manager

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
	"io"
	"os"
)

// Plays a song in DCA format
func (server *Server) playSound(el *queue.Element) (SkipReason, error) {
	var (
		opuslen    int16
		skip       bool
		skipReason SkipReason
		err        error
	)

	// Start speaking.
	_ = server.VC.SetSpeaking(true)
	audioChannel := server.VC.GetAudioChannel()

	go notify(notification.NotificationMessage{Notification: notification.Playing, Guild: server.GuildID})

	for {
		select {
		case <-server.Pause:
			go notify(notification.NotificationMessage{Notification: notification.Pause, Guild: server.GuildID})
			select {
			case <-server.Resume:
				go notify(notification.NotificationMessage{Notification: notification.Resume, Guild: server.GuildID})
				el.Segments = server.Queue.GetFirstElement().Segments
			case skipReason = <-server.Skip:
				cleanUp(server, el.Closer)
				return skipReason, nil
			}
		case skipReason = <-server.Skip:
			cleanUp(server, el.Closer)
			return skipReason, nil
		default:
			if el.Segments[int(server.Frames.Load())] {
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
					el.Reader = bufio.NewReader(f)
					el.Closer = f
					server.Frames.Store(0)
					go notify(notification.NotificationMessage{Notification: notification.LoopFinished, Guild: server.GuildID})
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
			server.Frames.Add(1)

			if skip {
				continue
			}

			// Should not be any end of file errors
			if err != nil {
				cleanUp(server, el.Closer)
				// Force to re-download the song
				_ = os.Remove(constants.CachePath + el.ID + constants.AudioExtension)
				return Error, err
			}

			audioChannel <- InBuf
		}
	}
}

func cleanUp(server *Server, closer io.Closer) {
	_ = server.VC.SetSpeaking(false)
	server.Frames.Store(0)

	if closer != nil {
		_ = closer.Close()
	}
}
