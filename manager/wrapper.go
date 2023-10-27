package manager

import (
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
	"strings"
	"time"
)

// Wrapper function for playing songs
func (server *Server) Play(clients *Clients, song string, i *discordgo.Interaction, guild, username string, random, loop, priority bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		server.spotifyPlaylist(clients, username, i, random, loop, priority, spotify.ID(strings.TrimPrefix(song, "spotify:playlist:")))
	case strings.Contains(song, "spotify.com/playlist/"):
		server.spotifyPlaylist(clients, username, i, random, loop, priority, spotify.ID(strings.Split(strings.TrimPrefix(song, "https://open.spotify.com/playlist/"), "?")[0]))
	case strings.HasPrefix(song, "spotify:track:"):
		server.spotifyTrack(clients, username, i, loop, priority, spotify.ID(strings.TrimPrefix(song, "spotify:track:")))
	case strings.Contains(song, "spotify.com/track/"):
		server.spotifyTrack(clients, username, i, loop, priority, spotify.ID(strings.Split(strings.TrimPrefix(song, "https://open.spotify.com/track/"), "?")[0]))
	case strings.HasPrefix(song, "spotify:album:"):
		server.spotifyAlbum(clients, username, i, random, loop, priority, spotify.ID(strings.TrimPrefix(song, "spotify:album:")))
	case strings.Contains(song, "spotify.com/album/"):
		server.spotifyAlbum(clients, username, i, random, loop, priority, spotify.ID(strings.Split(strings.TrimPrefix(song, "https://open.spotify.com/album/"), "?")[0]))
	case IsValidURL(song):
		server.downloadAndPlay(clients, song, username, i, random, loop, true, priority)
	default:
		link, err := searchDownloadAndPlay(song, clients.Youtube)
		if err == nil {
			server.downloadAndPlay(clients, link, username, i, random, loop, true, priority)
		} else {
			embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
		}
	}
}
