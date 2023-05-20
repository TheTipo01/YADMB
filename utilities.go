package main

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Finds user current voice channel
func findUserVoiceState(session *discordgo.Session, i *discordgo.Interaction) *discordgo.VoiceState {
	for _, guild := range session.State.Guilds {
		if guild.ID != i.GuildID {
			continue
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == i.Member.User.ID {
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

func sendEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) *discordgo.Message {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return nil
	}

	return m
}

func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string, wait time.Duration) {
	m := sendEmbed(s, embed, txtChannel)
	if m != nil {
		time.Sleep(wait)
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			lit.Error("ChannelMessageDelete failed: %s", err)
		}
	}

}

// Sends embed as response to an interaction
func sendEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c *chan int) {
	sliceEmbed := []*discordgo.MessageEmbed{embed}
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Embeds: sliceEmbed}})
	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		*c <- 1
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	sendEmbedInteraction(s, embed, i, nil)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Modify an already sent interaction
func modfyInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction) {
	sliceEmbed := []*discordgo.MessageEmbed{embed}
	_, err := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{Embeds: &sliceEmbed})
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}
}

// Modify an already sent interaction and deletes it after the specified wait time
func modfyInteractionAndDelete(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	modfyInteraction(s, embed, i)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Deletes an array of discordgo.Message
func deleteMessages(s *discordgo.Session, messages []discordgo.Message) {
	for _, m := range messages {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
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

	if server[guildID].queue.GetFirstElement() == nil && server[guildID].vc != nil {
		_ = server[guildID].vc.Disconnect()
		server[guildID].vc = nil
	}
}

// DirSize gets size of a directory
func DirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size
}

// ByteCountSI formats bytes into a readable format
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func deleteInteraction(s *discordgo.Session, i *discordgo.Interaction, c *chan int) {
	if c != nil {
		<-*c
	}
	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// idGen returns the first 11 characters of the SHA1 hash for the given link
func idGen(link string) string {
	h := sha1.New()
	h.Write([]byte(link))

	return strings.ToLower(base32.HexEncoding.EncodeToString(h.Sum(nil))[0:11])
}

func checkAudioOnly(formats RequestedFormats) bool {
	for _, f := range formats {
		if f.Resolution == "audio only" {
			return true
		}
	}

	return false
}

// isCommandNotAvailable checks whatever a command is available
func isCommandNotAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err != nil
}

// removePlaylist removes the playlist parameter from the url
func removePlaylist(s string) string {
	u, _ := url.Parse(s)
	q := u.Query()
	q.Del("list")
	u.RawQuery = q.Encode()
	return u.String()
}

func initializeServer(guild string) {
	if _, ok := server[guild]; !ok {
		server[guild] = NewServer(guild)
	}
}
