package manager

import (
	"net/url"
	"strings"
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/zmb3/spotify/v2"
)

// PlayEvent is the struct for playing songs
type PlayEvent struct {
	Username   string
	Song       string
	Clients    *Clients
	Event      *events.ApplicationCommandInteractionCreate
	Random     bool
	Loop       bool
	Priority   bool
	IsDeferred chan struct{}
}

// Play is a wrapper function for playing songs
func (server *Server) Play(p PlayEvent) {
	server.ChanQuitVC <- false

	if strings.Contains(p.Song, "spotify.com/") {
		// Parse URL
		u, err := url.Parse(p.Song)
		if err == nil {
			splitted := strings.Split(u.Path, "/")
			if len(splitted) >= 2 {
				if len(splitted) == 4 {
					// Remove second element
					splitted = append(splitted[:1], splitted[2:]...)
				}

				switch splitted[1] {
				case "track":
					server.spotifyTrack(p, spotify.ID(splitted[2]))
				case "playlist":
					server.spotifyPlaylist(p, spotify.ID(splitted[2]))
				case "album":
					server.spotifyAlbum(p, spotify.ID(splitted[2]))
				}

				return
			}
		}
	}

	if IsValidURL(p.Song) {
		server.downloadAndPlay(p, true)
	} else {
		var err error

		p.Song, err = searchDownloadAndPlay(p.Song, p.Clients.Youtube, p.Clients.Database)
		if err == nil {
			server.downloadAndPlay(p, true)
		} else {
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).
				AddField(constants.ErrorTitle, err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
		}
	}
}
