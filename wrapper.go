package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Wrapper function for playing songs
func play(s *discordgo.Session, song string, i *discordgo.Interaction, guild, username string, random bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, username, song, i, random)

	case isValidURL(song):
		downloadAndPlay(s, guild, song, username, i, random)

	default:
		searchDownloadAndPlay(s, guild, song, username, i, random)
	}
}
