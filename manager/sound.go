package manager

import (
	"encoding/binary"
	"errors"
	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
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
	_ = server.VC.Speaking(true)

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
			server.Frames++

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

			if server.VC != nil {
				server.VC.OpusSend <- InBuf
			} else {
				lit.Debug("VC is nil, triggering reconnection")
				server.VC, _ = server.Clients.Discord.ChannelVoiceJoin(server.GuildID, server.VoiceChannel, false, true)
			}
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
