package manager

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/disgoorg/disgo/voice"
)

// ErrVoiceLost is returned by the playback loop when the voice connection could
// not be recovered. The current song is dropped and playback stops so that the
// rest of the queue can be retried by the next play command.
var ErrVoiceLost = errors.New("voice connection lost")

// notifySync tells clients the exact playback position and whether audio is
// currently flowing. It is sent after a start, a pause/resume, a skip and
// whenever a voice interruption has been recovered.
func (server *Server) notifySync(frames int, playing bool) {
	notify(notification.NotificationMessage{
		Notification: notification.Sync,
		Guild:        server.GuildID,
		Frames:       frames,
		Playing:      &playing,
	})
}

// Plays a song in DCA format
func (server *Server) playSound(el *queue.Element) (SkipReason, error) {
	var (
		opuslen    int16
		skip       bool
		started    bool
		skipReason SkipReason
		err        error
	)

	conn, err := server.ensureVoiceUDP()
	if err != nil {
		return Error, err
	}

	// Start speaking.
	if err := server.SetSpeaking(voice.SpeakingFlagMicrophone); err != nil {
		conn, err = server.reconnectVoice()
		if err != nil {
			return Error, err
		}
	}

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	notify(notification.NotificationMessage{Notification: notification.Playing, Guild: server.GuildID})

	for {
		select {
		case <-server.Pause:
			notify(notification.NotificationMessage{Notification: notification.Pause, Guild: server.GuildID})
			server.notifySync(int(server.Frames.Load()), false)
			select {
			case <-server.Resume:
				notify(notification.NotificationMessage{Notification: notification.Resume, Guild: server.GuildID})
				if started {
					server.notifySync(int(server.Frames.Load()), true)
				}
			case target := <-server.Seek:
				seekToFrame(el, target, int(server.Frames.Load()))
			case skipReason = <-server.Skip:
				cleanUp(server, el.Closer)
				return skipReason, nil
			}
		case target := <-server.Seek:
			seekToFrame(el, target, int(server.Frames.Load()))
		case skipReason = <-server.Skip:
			cleanUp(server, el.Closer)
			return skipReason, nil
		default:
			frame := int(server.Frames.Load())
			if _, ok := el.Segments[frame]; ok {
				skip = !skip

				// The target frame was reached: resync clients instead of
				// letting their local clock drift during the fast-forward.
				if !skip {
					server.notifySync(frame, true)
				}
			}

			// Read opus frame length from dca file.
			err = binary.Read(el.Reader, binary.LittleEndian, &opuslen)

			// If we got an actual read error, bail out, removing the partial file
			// so that it gets downloaded again.
			if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				cleanUp(server, el.Closer)

				// A clear operation closes the reader on purpose. Don't treat the
				// resulting error as a broken cache file.
				if server.Clear.Load() || server.cleaning.Load() {
					return Clear, nil
				}

				_ = os.Remove(constants.CachePath + el.ID + constants.AudioExtension)
				return Error, err
			}

			// If this is the end of the file, just return.
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				if server.Clear.Load() || server.cleaning.Load() {
					cleanUp(server, el.Closer)
					return Clear, nil
				}

				var pipeErrors strings.Builder

				for _, stderr := range el.Errors {
					bytes, err := io.ReadAll(stderr)
					if err == nil && len(bytes) > 0 {
						pipeErrors.WriteString(string(bytes))
					}
				}

				// If we read something from the stderr pipe, there's an error and we need to report it
				if finalString := pipeErrors.String(); finalString != "" {
					cleanUp(server, el.Closer)
					return Error, errors.New(finalString)
				}

				// Else, the file just finished normally
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
					notify(notification.NotificationMessage{Notification: notification.LoopFinished, Guild: server.GuildID})
					continue
				} else {
					cleanUp(server, el.Closer)
					return Finished, nil
				}
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opuslen)
			err = binary.Read(el.Reader, binary.LittleEndian, &InBuf)

			// Should not be any end of file errors
			if err != nil {
				cleanUp(server, el.Closer)

				if server.Clear.Load() || server.cleaning.Load() {
					return Clear, nil
				}

				// Force to re-download the song
				_ = os.Remove(constants.CachePath + el.ID + constants.AudioExtension)
				return Error, err
			}

			// The first successfully read frame means the source is actually
			// producing audio. Sync clients to the real playback position.
			if !started {
				started = true
				server.notifySync(int(server.Frames.Load()), true)
			}

			if skip {
				// Keep count of the frames in the song even when we don't send them.
				server.Frames.Add(1)
				continue
			}

			conn, err = server.writeFrame(conn, InBuf)
			if err != nil {
				cleanUp(server, el.Closer)
				return Error, err
			}

			// Keep count of the frames in the song
			server.Frames.Add(1)

			<-ticker.C
		}
	}
}

// SetSpeaking changes the bot speaking state, translating a missing connection
// into the vc error instead of a nil pointer dereference.
func (server *Server) SetSpeaking(flag voice.SpeakingFlags) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.VC.SetSpeaking(ctx, flag)
}

// ensureVoiceUDP returns a usable UDP connection, letting disgo finish an
// in-progress reconnect before forcing a new connection.
func (server *Server) ensureVoiceUDP() (voice.UDPConn, error) {
	conn, err := server.VC.UDP()
	if err == nil {
		return conn, nil
	}

	if conn, err := server.waitForVoice(nil); err == nil {
		return conn, nil
	}

	return server.reconnectVoice()
}

// reconnectVoice tries to reopen the voice connection and returns the new UDP
// connection. It is used both for resume-after-write-error and for a connection
// that disappeared between songs.
func (server *Server) reconnectVoice() (voice.UDPConn, error) {
	if server.Clear.Load() {
		return nil, ErrVoiceLost
	}

	if server.Clients == nil || server.Clients.Discord == nil {
		return nil, ErrVoiceLost
	}

	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		err := server.VC.Reconnect(ctx, server.Clients.Discord.VoiceManager)
		if err == nil {
			err = server.VC.SetSpeaking(ctx, voice.SpeakingFlagMicrophone)
		}

		var conn voice.UDPConn
		if err == nil {
			conn, err = server.VC.UDP()
		}

		cancel()

		if err == nil {
			return conn, nil
		}

		slog.Warn("voice reconnect failed", "guild", server.GuildID, "attempt", attempt+1, "error", err)

		if server.Clear.Load() {
			return nil, ErrVoiceLost
		}

		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	return nil, ErrVoiceLost
}

// writeFrame writes a single opus frame. If the UDP connection is gone, it
// first gives disgo a chance to recover the existing voice connection (for
// example after the bot was moved to another channel), then falls back to a
// forced reconnect. The same frame is retried so playback resumes in place.
func (server *Server) writeFrame(conn voice.UDPConn, frame []byte) (voice.UDPConn, error) {
	if _, err := conn.Write(frame); err == nil {
		return conn, nil
	} else {
		slog.Warn("voice write failed, trying to recover", "guild", server.GuildID, "error", err)
	}

	// Audio stopped flowing at this position. Freeze client clocks until the
	// connection is usable again.
	server.notifySync(int(server.Frames.Load()), false)

	if newConn, err := server.waitForVoice(frame); err == nil {
		server.notifySync(int(server.Frames.Load()), true)
		return newConn, nil
	}

	newConn, err := server.reconnectVoice()
	if err != nil {
		return nil, err
	}

	if _, err := newConn.Write(frame); err != nil {
		slog.Warn("voice write failed after reconnect", "guild", server.GuildID, "error", err)
		return nil, ErrVoiceLost
	}

	server.notifySync(int(server.Frames.Load()), true)

	return newConn, nil
}

// waitForVoice waits for disgo's own reconnect logic to make the voice
// connection usable again. This is the common path when Discord moves the bot
// between channels: the old call is terminated and disgo opens a fresh voice
// session. Retrying the same frame avoids losing the current song.
func (server *Server) waitForVoice(frame []byte) (voice.UDPConn, error) {
	const (
		maxWait           = 35 * time.Second
		pollInterval      = 250 * time.Millisecond
		disconnectedGrace = 3 * time.Second
	)

	deadline := time.Now().Add(maxWait)
	var disconnectedSince time.Time

	for time.Now().Before(deadline) {
		if server.Clear.Load() {
			return nil, ErrVoiceLost
		}

		switch server.VC.GatewayStatus() {
		case voice.StatusDisconnected, voice.StatusUnconnected:
			if disconnectedSince.IsZero() {
				disconnectedSince = time.Now()
			} else if time.Since(disconnectedSince) >= disconnectedGrace {
				// disgo is not reconnecting; let the caller force a fresh conn.
				return nil, ErrVoiceLost
			}
		default:
			disconnectedSince = time.Time{}

			if conn, err := server.VC.UDP(); err == nil {
				if frame == nil {
					return conn, nil
				}

				if _, err := conn.Write(frame); err == nil {
					return conn, nil
				}
			}
		}

		time.Sleep(pollInterval)
	}

	return nil, ErrVoiceLost
}

func seekToFrame(el *queue.Element, target, current int) {
	if target <= current+1 {
		return
	}

	if el.Segments == nil {
		el.Segments = make(map[int]struct{}, 2)
	}

	el.Segments[current+1] = struct{}{}
	el.Segments[target] = struct{}{}
}

func cleanUp(server *Server, closer io.Closer) {
	_ = server.SetSpeaking(voice.SpeakingFlagNone)
	server.Frames.Store(0)

	if closer != nil {
		_ = closer.Close()
	}
}
