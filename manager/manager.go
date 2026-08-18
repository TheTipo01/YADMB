package manager

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/vc"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

var (
	Notifications = make(chan notification.NotificationMessage, 1)
	BotName       string
)

// NewServer creates a new server manager
func NewServer(guildID snowflake.ID, clients *Clients) *Server {
	var server = &Server{
		Queue:      queue.NewQueue(),
		Custom:     make(map[string]*database.CustomCommand),
		GuildID:    guildID,
		Pause:      make(chan struct{}),
		Resume:     make(chan struct{}),
		Skip:       make(chan SkipReason),
		Started:    atomic.Bool{},
		Clear:      atomic.Bool{},
		Paused:     atomic.Bool{},
		WG:         &sync.WaitGroup{},
		Clients:    clients,
		VC:         vc.NewVC(guildID),
		ChanQuitVC: make(chan bool),
	}

	go server.handleQuitVC()

	return server
}

// AddSong adds a song to the queue
func (server *Server) AddSong(priority bool, el ...queue.Element) {
	if priority {
		go notify(notification.NotificationMessage{Notification: notification.PrioritySong, Songs: el, Guild: server.GuildID})
	} else {
		go notify(notification.NotificationMessage{Notification: notification.NewSongs, Songs: el, Guild: server.GuildID})
	}

	if priority {
		server.Queue.AddElementsPriority(el...)
	} else {
		server.Queue.AddElements(el...)
	}

	if server.Started.CompareAndSwap(false, true) {
		go server.play()
	}
}

func (server *Server) play() {
	msg := make(chan *discord.Message)

	server.Paused.Store(false)

	for el := server.Queue.GetFirstElement(); el != nil && !server.Clear.Load(); el = server.Queue.GetFirstElement() {
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

		if err == nil {
			skipReason, err := server.playSound(el)
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
				err := el.AfterPlay()
				if err == nil {
					go func() {
						embed.SendAndDeleteEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
							AddField(constants.ErrorTitle, err.Error(), false).
							WithColor(0x7289DA), el.TextChannel, time.Second*15)
					}()
				}
			}

			if err != nil {
				embed.SendAndDeleteEmbed(server.Clients.Discord, discord.NewEmbed().WithTitle(BotName).
					AddField(constants.ErrorTitle, err.Error(), false).
					WithColor(0x7289DA), el.TextChannel, time.Second*15)
			}

			if skipReason == Finished {
				go notify(notification.NotificationMessage{Notification: notification.Finished, Guild: server.GuildID})
			} else {
				go notify(notification.NotificationMessage{Notification: notification.Skip, Guild: server.GuildID})
			}
		} else {
			go waitMessageAndDelete(msg, server.Clients.Discord)
		}

		server.Queue.RemoveFirstElement()
	}

	server.Started.Store(false)

	server.ChanQuitVC <- true
}

// IsPlaying returns whether the bot is playing
func (server *Server) IsPlaying() bool {
	return server.Started.Load() && !server.Queue.IsEmpty()
}

// Clean clears the queue
func (server *Server) Clean() {
	if server.IsPlaying() {
		server.Clear.Store(true)
		server.Skip <- Clear

		go notify(notification.NotificationMessage{Notification: notification.Clear, Guild: server.GuildID})

		server.WG.Wait()
		server.Clear.Store(false)

		q := server.Queue.GetAllQueue()
		server.Queue.Clear()

		for _, el := range q {
			if el.Closer != nil {
				_ = el.Closer.Close()
			}
		}

		server.ChanQuitVC <- true
	}
}

func (server *Server) handleQuitVC() {
	var c bool
	var timer *time.Timer

	for {
		// Wait for a signal in the channel
		c = <-server.ChanQuitVC
		if c {
			if timer == nil {
				timer = time.AfterFunc(time.Minute, server.QuitVC)
			}
		} else {
			if timer != nil {
				timer.Stop()
				timer = nil
			}
		}
	}
}
