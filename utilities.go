package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

// Logs and instantly delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	lit.Debug(m.Author.Username + ": " + m.Content)

	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		lit.Error("Can't delete message, %s", err)
	}
}

// Finds user current voice channel
func findUserVoiceState(session *discordgo.Session, m *discordgo.MessageCreate) *discordgo.VoiceState {

	for _, guild := range session.State.Guilds {
		if guild.ID != m.GuildID {
			continue
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == m.Author.ID {
				return vs
			}
		}
	}

	return nil
}

// Checks if a string is a valid URL
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

// Removes element from the queue
func removeFromQueue(id string, guild string) {
	for i, q := range server[guild].queue {
		if q.id == id {
			copy(server[guild].queue[i:], server[guild].queue[i+1:])
			server[guild].queue[len(server[guild].queue)-1] = Queue{"", "", "", "", "", nil, 0, "", nil}
			server[guild].queue = server[guild].queue[:len(server[guild].queue)-1]
			return
		}
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return
	}

	time.Sleep(time.Second * 5)

	err = s.ChannelMessageDelete(txtChannel, m.ID)
	if err != nil {
		lit.Error("MessageDelete failed: %s", err)
		return
	}
}

// Formats a string given it's duration in seconds
func formatDuration(duration float64) string {
	duration2 := int(duration)
	hours := duration2 / 3600
	duration2 -= 3600 * hours
	minutes := (duration2) / 60
	duration2 -= minutes * 60

	if hours != 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, duration2)
	}

	if minutes != 0 {
		return fmt.Sprintf("%02d:%02d", minutes, duration2)
	}

	return fmt.Sprintf("%02d", duration2)

}

// Split lyrics into smaller messages
func formatLongMessage(text []string) []string {
	var counter int
	var output []string
	var buffer string
	const charLimit = 1900

	for _, line := range text {
		counter += strings.Count(line, "")

		// If the counter is exceeded, we append all the current line to the final slice
		if counter > charLimit {
			counter = 0
			output = append(output, buffer)

			buffer = line + "\n"
			continue
		}

		buffer += line + "\n"

	}

	return append(output, buffer)
}

// Deletes an array of discordgo.Message
func deleteMessages(s *discordgo.Session, messages []discordgo.Message) {
	for _, m := range messages {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

// Shuffles a slice of strings
func shuffle(a []string) []string {
	final := make([]string, len(a))

	for i, v := range rand.Perm(len(a)) {
		final[v] = a[i]
	}
	return final
}

// Disconnects the bot from the voice channel after 1 minute if nothing is playing
func quitVC(guildID string) {
	time.Sleep(1 * time.Minute)

	if len(server[guildID].queue) == 0 && server[guildID].vc != nil {
		server[guildID].server.Lock()

		_ = server[guildID].vc.Disconnect()
		server[guildID].vc = nil

		server[guildID].server.Unlock()
	}
}

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
