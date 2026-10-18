package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/vc"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

var (
	Notifications = make(chan notification.NotificationMessage, 64)
	BotName       string
)

// NewServer creates a new server manager
func NewServer(guildID snowflake.ID, clients *Clients) *Server {
	return &Server{
		Queue:   queue.NewQueue(),
		GuildID: guildID,
		Pause:   make(chan struct{}, 1),
		Resume:  make(chan struct{}, 1),
		Skip:    make(chan SkipReason, 1),
		Seek:    make(chan int, 1),
		Started: atomic.Bool{},
		Clear:   atomic.Bool{},
		Paused:  atomic.Bool{},
		WG:      &sync.WaitGroup{},
		Clients: clients,
		VC:      vc.NewVC(guildID),
	}
}

// AddSong adds a song to the queue
func (server *Server) AddSong(priority bool, el ...queue.Element) {
	if server.shuttingDown.Load() {
		for _, element := range el {
			if element.Closer != nil {
				_ = element.Closer.Close()
			}
		}
		return
	}

	if priority {
		notify(notification.NotificationMessage{Notification: notification.PrioritySong, Songs: el, Guild: server.GuildID})
	} else {
		notify(notification.NotificationMessage{Notification: notification.NewSongs, Songs: el, Guild: server.GuildID})
	}

	server.cleanMu.Lock()
	defer server.cleanMu.Unlock()

	if priority {
		server.Queue.AddElementsPriority(el...)
	} else {
		server.Queue.AddElements(el...)
	}

	server.startPlayback()
}

// startPlayback starts the player if there is work to do. It is safe to call
// concurrently: the Started flag and the player goroutine are guarded by playMu.
func (server *Server) startPlayback() {
	server.playMu.Lock()
	defer server.playMu.Unlock()

	server.startPlaybackLocked()
}

func (server *Server) startPlaybackLocked() {
	server.stopQuitLocked()

	if server.shuttingDown.Load() || server.Started.Load() || server.Queue.IsEmpty() || server.Clear.Load() {
		return
	}

	server.Started.Store(true)
	server.Paused.Store(false)
	drain(server.Skip)
	drain(server.Pause)
	drain(server.Resume)
	drain(server.Seek)

	done := make(chan struct{})
	server.playDone = done

	go server.play(done)
}

func (server *Server) play(done chan struct{}) {
	defer close(done)

	restart := true
	allowQuit := true

	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from playback panic", "guild", server.GuildID, "error", r)
			restart = false
		}

		server.finishPlayback(restart, allowQuit)
	}()

	msg := make(chan *discord.Message)

	server.Paused.Store(false)

	for {
		if server.Clear.Load() {
			restart = false
			break
		}

		// The playing element is a copy returned by GetFirstElement. Make sure
		// its segments map is shared with the real queue element, otherwise a
		// /goto on a song without SponsorBlock segments would be invisible.
		server.Queue.ModifyFirstElement(func(el *queue.Element) {
			if el.Segments == nil {
				el.Segments = make(map[int]struct{}, 2)
			}
		})

		el := server.Queue.GetFirstElement()
		if el == nil {
			break
		}

		// Send "Now playing" message
		go func() {
			msg <- embed.SendEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
				AddField("Now playing", fmt.Sprintf("[%s](%s) - %s added by %s", el.Title,
					el.Link, el.Duration, el.User), false).
				WithColor(0x7289DA).WithThumbnail(el.Thumbnail), el.TextChannel)
		}()

		var err error

		if el.BeforePlay != nil {
			err = el.BeforePlay()
		}

		if err != nil {
			go waitMessageAndDelete(msg, server.Clients.Discord)
			server.Queue.RemoveFirstElement()
			continue
		}

		skipReason, playErr := server.playSound(el)
		go waitMessageAndDelete(msg, server.Clients.Discord)

		// If we are still downloading the song, we need to finish writing it to disk
		if el.Downloading && skipReason > Finished {
			// Notify the user
			go func() {
				msg <- embed.SendEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
					AddField(constants.ProcessingTitle, constants.Processing, false).
					WithColor(0x7289DA), el.TextChannel)
			}()

			devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
			_, _ = io.Copy(devnull, el.Reader)
			_ = devnull.Close()

			go waitMessageAndDelete(msg, server.Clients.Discord)
		}

		if el.AfterPlay != nil {
			if err := el.AfterPlay(); err != nil {
				go func() {
					embed.SendAndDeleteEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
						AddField(constants.ErrorTitle, err.Error(), false).
						WithColor(0x7289DA), el.TextChannel, time.Second*15)
				}()
			}
		}

		if playErr != nil {
			go embed.SendAndDeleteEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
				AddField(constants.ErrorTitle, playErr.Error(), false).
				WithColor(0x7289DA), el.TextChannel, time.Second*15)
		}

		if skipReason == Finished {
			notify(notification.NotificationMessage{Notification: notification.Finished, Guild: server.GuildID})
		} else {
			notify(notification.NotificationMessage{Notification: notification.Skip, Guild: server.GuildID})
		}

		// The reader of the current element is partially consumed. Drop it and
		// stop instead of trying to restart it from an unknown position.
		if errors.Is(playErr, ErrVoiceLost) {
			server.Queue.RemoveFirstElement()
			restart = false
			break
		}

		server.Queue.RemoveFirstElement()
	}

	if server.Clear.Load() {
		restart = false
	}
}

func (server *Server) finishPlayback(restart, allowQuit bool) {
	server.playMu.Lock()
	defer server.playMu.Unlock()

	server.Started.Store(false)
	server.playDone = nil

	if restart && !server.Clear.Load() && !server.Queue.IsEmpty() {
		server.startPlaybackLocked()
		return
	}

	if allowQuit && !server.Clear.Load() && server.Queue.IsEmpty() {
		server.scheduleQuitLocked()
	}
}

func (server *Server) PausePlayback() bool {
	if !server.Paused.CompareAndSwap(false, true) {
		return false
	}

	select {
	case server.Pause <- struct{}{}:
	default:
	}

	return true
}

func (server *Server) ResumePlayback() bool {
	if !server.Paused.CompareAndSwap(true, false) {
		return false
	}

	select {
	case server.Resume <- struct{}{}:
	default:
	}

	return true
}

func (server *Server) SkipSong(reason SkipReason) bool {
	select {
	case server.Skip <- reason:
	default:
	}

	return true
}

// SeekFrame asks the player to skip to the given frame.
func (server *Server) SeekFrame(frame int) bool {
	select {
	case server.Seek <- frame:
		return true
	default:
		return false
	}
}

// IsPlaying returns whether the bot is playing
func (server *Server) IsPlaying() bool {
	return server.Started.Load() && !server.Queue.IsEmpty()
}

// CancelQuit cancels a pending idle disconnect.
func (server *Server) CancelQuit() {
	server.playMu.Lock()
	defer server.playMu.Unlock()

	server.stopQuitLocked()
}

// Clean clears the queue
func (server *Server) Clean() {
	server.clean(context.Background(), true, false)
}

func (server *Server) clean(ctx context.Context, waitDownloads, shutdown bool) {
	if !server.cleaning.CompareAndSwap(false, true) {
		return
	}
	defer server.cleaning.Store(false)

	server.Clear.Store(true)

	server.playMu.Lock()
	done := server.playDone
	server.playMu.Unlock()

	if done != nil {
		select {
		case server.Skip <- Clear:
		default:
		}

		select {
		case <-done:
		case <-ctx.Done():
		case <-time.After(10 * time.Second):
			// Closing the current reader unblocks a player stuck on a slow
			// stream. The player checks Clear and exits without removing the
			// cache file.
			for _, el := range server.Queue.GetAllQueue() {
				if el.Closer != nil {
					_ = el.Closer.Close()
				}
			}

			select {
			case <-done:
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
		}
	}

	// Wait for playlist downloads to notice the Clear flag.
	if waitDownloads {
		server.WG.Wait()
	} else {
		doneWG := make(chan struct{})
		go func() {
			server.WG.Wait()
			close(doneWG)
		}()

		select {
		case <-doneWG:
		case <-ctx.Done():
		}
	}

	server.cleanMu.Lock()
	q := server.Queue.GetAllQueue()
	server.Queue.Clear()

	for _, el := range q {
		if el.Closer != nil {
			_ = el.Closer.Close()
		}
	}

	server.Clear.Store(false)
	server.cleanMu.Unlock()

	notify(notification.NotificationMessage{Notification: notification.Clear, Guild: server.GuildID})

	if shutdown {
		return
	}

	server.playMu.Lock()
	if !server.Started.Load() && server.Queue.IsEmpty() {
		server.scheduleQuitLocked()
	}
	server.playMu.Unlock()
}

func (server *Server) scheduleQuitLocked() {
	if server.quitTimer == nil {
		server.quitTimer = time.AfterFunc(time.Minute, server.quitIfIdle)
		return
	}

	server.quitTimer.Reset(time.Minute)
}

func (server *Server) stopQuitLocked() {
	if server.quitTimer != nil {
		server.quitTimer.Stop()
		server.quitTimer = nil
	}
}

func (server *Server) quitIfIdle() {
	server.playMu.Lock()
	defer server.playMu.Unlock()

	server.quitTimer = nil

	if server.Started.Load() || !server.Queue.IsEmpty() || server.Clear.Load() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.VC.Close(ctx)
}

func drain[T any](ch chan T) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
