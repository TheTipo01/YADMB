package manager

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
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
	err = server.VC.SetSpeaking(true)
	if err != nil {
		return Error, err
	}
	audioChannel := server.VC.GetAudioChannel()
	dead := server.VC.GetIsDeadChannel()

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
			if _, ok := el.Segments[int(server.Frames.Load())]; ok {
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

					f, err := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
					if err != nil {
						return Error, err
					}
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

			// TODO: does not work, i don't understand if it's my fault or the library's
			select {
			case <-dead:
				cleanUp(server, el.Closer)
				return Error, errors.New("voice connection dead")
			default:
				audioChannel <- InBuf
			}
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
