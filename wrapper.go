package main

import (
	"github.com/bwmarrin/discordgo"
	"io"
	"os/exec"
	"strings"
)

// Wrapper function for playing songs
func play(s *discordgo.Session, song, textChannel, voiceChannel, guild, username string, random bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, voiceChannel, username, song, textChannel, random)
		break

	case isValidURL(song):
		downloadAndPlay(s, guild, voiceChannel, song, username, textChannel, random)
		break

	default:
		searchDownloadAndPlay(s, guild, voiceChannel, song, username, textChannel, random)
	}
}

// Wrapper function for soundStream, also waits for the song to finish to download and then closes it's pipe
func playSoundStream(s *discordgo.Session, guildID, channelID, fileName, txtChannel string, stdout io.ReadCloser, cmd *exec.Cmd) {

	soundStream(s, guildID, channelID, fileName, txtChannel, stdout)

	_ = stdout.Close()
	_ = cmd.Wait()

}
