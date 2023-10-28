package manager

import (
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
	"strings"
	"time"
)

// PlayEvent is the struct for playing songs
type PlayEvent struct {
	Username    string
	Song        string
	Clients     *Clients
	Interaction *discordgo.Interaction
	Random      bool
	Loop        bool
	Priority    bool
	IsDeferred  chan struct{}
}

// Wrapper function for playing songs
func (server *Server) Play(p PlayEvent) {
	switch {
	case strings.HasPrefix(p.Song, "spotify:playlist:"):
		server.spotifyPlaylist(p, spotify.ID(strings.TrimPrefix(p.Song, "spotify:playlist:")))
	case strings.Contains(p.Song, "spotify.com/playlist/"):
		server.spotifyPlaylist(p, spotify.ID(strings.Split(strings.TrimPrefix(p.Song, "https://open.spotify.com/playlist/"), "?")[0]))
	case strings.HasPrefix(p.Song, "spotify:track:"):
		server.spotifyTrack(p, spotify.ID(strings.TrimPrefix(p.Song, "spotify:track:")))
	case strings.Contains(p.Song, "spotify.com/track/"):
		server.spotifyTrack(p, spotify.ID(strings.Split(strings.TrimPrefix(p.Song, "https://open.spotify.com/track/"), "?")[0]))
	case strings.HasPrefix(p.Song, "spotify:album:"):
		server.spotifyAlbum(p, spotify.ID(strings.TrimPrefix(p.Song, "spotify:album:")))
	case strings.Contains(p.Song, "spotify.com/album/"):
		server.spotifyAlbum(p, spotify.ID(strings.Split(strings.TrimPrefix(p.Song, "https://open.spotify.com/album/"), "?")[0]))
	case IsValidURL(p.Song):
		server.downloadAndPlay(p, true)
	default:
		var err error

		p.Song, err = searchDownloadAndPlay(p.Song, p.Clients.Youtube)
		if err == nil {
			server.downloadAndPlay(p, true)
		} else {
			embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).
				AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
		}
	}
}
