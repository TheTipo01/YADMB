package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
	"strings"
	"time"
)

// Wrapper function for playing songs
func play(s *discordgo.Session, song string, i *discordgo.Interaction, guild, username string, random, loop bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, username, i, random, loop, spotify.ID(strings.TrimPrefix(song, "spotify:playlist:")))
	case strings.Contains(song, "spotify.com/playlist/"):
		spotifyPlaylist(s, guild, username, i, random, loop, spotify.ID(strings.Split(strings.TrimPrefix(song, "https://open.spotify.com/playlist/"), "?")[0]))
	case isValidURL(song):
		downloadAndPlay(s, guild, song, username, i, random, loop, true)
	default:
		link, err := searchDownloadAndPlay(song)
		if err == nil {
			downloadAndPlay(s, guild, link, username, i, random, loop, true)
		} else {
			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		}
	}
}
