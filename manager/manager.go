package manager

import (
	"fmt"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// NewServer creates a new server manager
func NewServer(guildID string, clients Clients) *Server {
	return &Server{
		Queue:               queue.NewQueue(),
		Custom:              make(map[string]*database.CustomCommand),
		GuildID:             guildID,
		Pause:               make(chan struct{}),
		Resume:              make(chan struct{}),
		Skip:                make(chan SkipReason),
		Started:             atomic.Bool{},
		Clear:               atomic.Bool{},
		Paused:              atomic.Bool{},
		WG:                  &sync.WaitGroup{},
		VoiceChannelMembers: make(map[string]*atomic.Int32),
		Clients:             clients,
	}
}

// AddSong adds a song to the queue
func (server *Server) AddSong(priority bool, el ...queue.Element) {
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
	msg := make(chan *discordgo.Message)

	server.Paused.Store(false)

	for el := server.Queue.GetFirstElement(); el != nil && !server.Clear.Load(); el = server.Queue.GetFirstElement() {
		// Send "Now playing" message
		go func() {
			msg <- embed.SendEmbed(server.Clients.Discord, embed.NewEmbed().SetTitle(server.Clients.Discord.State.User.Username).
				AddField("Now playing", fmt.Sprintf("[%s](%s) - %s added by %s", el.Title,
					el.Link, el.Duration, el.User)).
				SetColor(0x7289DA).SetThumbnail(el.Thumbnail).MessageEmbed, el.TextChannel)
		}()

		if el.BeforePlay != nil {
			el.BeforePlay()
		}

		skipReason, _ := playSound(el, server)

		// If we are still downloading the song, we need to finish writing it to disk
		if el.Downloading && skipReason > Finished {
			devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
			_, _ = io.Copy(devnull, el.Reader)
			_ = devnull.Close()
		}

		if el.AfterPlay != nil {
			el.AfterPlay()
		}

		// Delete it after it has been played
		go func() {
			if message := <-msg; message != nil {
				_ = server.Clients.Discord.ChannelMessageDelete(message.ChannelID, message.ID)
			}
		}()

		if skipReason != Clear {
			server.Queue.RemoveFirstElement()
		}
	}

	server.Started.Store(false)

	go QuitVC(server)
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

		server.WG.Wait()
		server.Clear.Store(false)

		q := server.Queue.GetAllQueue()
		server.Queue.Clear()

		for _, el := range q {
			if el.Closer != nil {
				_ = el.Closer.Close()
			}
		}
	}
}
