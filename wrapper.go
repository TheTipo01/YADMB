package main

import (
	"encoding/binary"
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
	var (
		err     error
		opuslen int16
	)

	soundStream(s, guildID, channelID, fileName, txtChannel, stdout)

	// TODO: Maybe I can find another way to wait for the song to finish to download?
	for {
		// Read opus frame length from dca file.
		err = binary.Read(stdout, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		InBuf := make([]byte, opuslen)
		err = binary.Read(stdout, binary.LittleEndian, InBuf)

	}

	_ = stdout.Close()
	_ = cmd.Wait()

}
