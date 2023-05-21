package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Wrapper function for playing songs
func play(s *discordgo.Session, song string, i *discordgo.Interaction, guild, username string, random, loop bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, username, song, i, random, loop)

	case isValidURL(song):
		downloadAndPlay(s, guild, song, username, i, random, loop)

	default:
		searchDownloadAndPlay(s, guild, song, username, i, random, loop)
	}
}
