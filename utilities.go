package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
			server[guild].queue[len(server[guild].queue)-1] = Queue{"", "", "", "", "", nil, "", 0, nil}
			server[guild].queue = server[guild].queue[:len(server[guild].queue)-1]
			return
		}
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string, wait time.Duration) {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return
	}

	time.Sleep(wait)

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

// Returns a map for skipping certain frames of a song
func getSegments(videoID string) map[int]bool {

	// Gets segments
	resp, err := http.Get("https://sponsor.ajay.app/api/skipSegments?videoID=" + videoID + "&categories=[\"sponsor\",\"music_offtopic\"]")
	if err != nil {
		lit.Error("Can't get SponsorBlock segments", err)
		return nil
	}

	// If we get the HTTP code 200, segments were found for the given video
	if resp.StatusCode == http.StatusOK {
		var segments SponsorBlock
		segmentMap := make(map[int]bool)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lit.Error("Can't read response body", err)
			return nil
		}

		err = resp.Body.Close()
		if err != nil {
			lit.Error("Can't close response body", err)
			return nil
		}

		err = json.Unmarshal(body, &segments)
		if err != nil {
			lit.Error("Can't unmarshal JSON", err)
			return nil
		}

		for _, s := range segments {
			if len(s.Segment) == 2 {
				segmentMap[int(s.Segment[0]*frameSeconds)] = true
				segmentMap[int(s.Segment[1]*frameSeconds)] = true
			}
		}

		return segmentMap
	}

	return nil

}

// From a map of segments returns an encoded string
func encodeSegments(segments map[int]bool) string {
	if segments == nil {
		return ""
	}

	var out string

	for k := range segments {
		out += strconv.Itoa(k) + ","
	}

	return strings.TrimSuffix(out, ",")
}

// Decodes segments into a map
func decodeSegments(segments string) map[int]bool {
	if segments == "" {
		return nil
	}

	mapSegments := make(map[int]bool)
	splitted := strings.Split(segments, ",")

	for _, s := range splitted {
		frame, err := strconv.Atoi(s)
		if err == nil {
			mapSegments[frame] = true
		}
	}

	return mapSegments
}
