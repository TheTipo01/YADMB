package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os/exec"
	"runtime"
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

	switch runtime.GOOS {
	case "windows":
		_ = stdout.Close()
		break
		// TODO: On linux, if we close the pipe, tee just quits without waiting for the song to completely download
	default:
		var err error
		var opuslen int16

		for {
			// Read opus frame length from dca file.
			err = binary.Read(stdout, binary.LittleEndian, &opuslen)

			// If this is the end of the file, just return.
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}

			if err != nil {
				lit.Error("Error reading from dca file: %s", err)
				break
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opuslen)
			err = binary.Read(stdout, binary.LittleEndian, &InBuf)
		}
	}

	_ = cmd.Wait()

}
