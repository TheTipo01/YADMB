package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Wrapper function for playing songs
func play(s *discordgo.Session, song string, i *discordgo.Interaction, voiceChannel, guild, username string, random bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, voiceChannel, username, song, i, random)

	case isValidURL(song):
		downloadAndPlay(s, guild, voiceChannel, song, username, i, random)

	default:
		searchDownloadAndPlay(s, guild, voiceChannel, song, username, i, random)
	}
}
