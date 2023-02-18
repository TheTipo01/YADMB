package main

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

// Removes element from the queue
func removeFromQueue(id string, guild string) {
	server[guild].queueMutex.Lock()
	defer server[guild].queueMutex.Unlock()
	for i, q := range server[guild].queue {
		if q.id == id {
			copy(server[guild].queue[i:], server[guild].queue[i+1:])
			server[guild].queue[len(server[guild].queue)-1] = Queue{"", "", "", "", "", nil, "", 0, nil, "", ""}
			server[guild].queue = server[guild].queue[:len(server[guild].queue)-1]
			return
		}
	}
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

// If server is nil, tries to initialize it
func initializeServer(guild string) {
	if server[guild] == nil {
		server[guild] = &Server{server: &sync.Mutex{}, pause: &sync.Mutex{}, stream: &sync.Mutex{}, queueMutex: &sync.Mutex{}, custom: make(map[string]*CustomCommand)}
	}
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

// Downloads a song (if it's not cached), deletes an interaction response, and return a prepared Queue element
func downloadSong(s *discordgo.Session, i *discordgo.Interaction, url string, c *chan int, channelID string) (*Queue, error) {
	// Check if the song is the db, to speedup things
	el := checkInDb(url)

	info, err := os.Stat(cachePath + el.id + audioExtension)
	// Not in db, download it
	if el.title == "" || err != nil || info.Size() <= 0 {
		// Get info about it
		splittedOut, err := getInfo(url)
		if err != nil {
			return nil, err
		}

		var ytdl YtDLP
		_ = json.Unmarshal([]byte(splittedOut[0]), &ytdl)

		cmds := download(url, checkAudioOnly(ytdl.RequestedFormats))

		el = Queue{ytdl.Title, formatDuration(ytdl.Duration), "", ytdl.WebpageURL, i.Member.User.Username, nil, ytdl.Thumbnail, 0, nil, channelID, i.ChannelID}

		exist := false
		switch ytdl.Extractor {
		case "youtube":
			el.id = ytdl.ID + "-" + ytdl.Extractor
			// SponsorBlock is supported only on youtube
			el.segments = getSegments(ytdl.ID)

			// If the song is on YouTube, we also add it with its compact url, for faster parsing
			addToDb(el, false)
			exist = true

			el.link = "https://youtu.be/" + ytdl.ID
		case "generic":
			// The generic extractor doesn't give out something unique, so we generate one from the link
			el.id = idGen(el.link) + "-" + ytdl.Extractor
		default:
			el.id = ytdl.ID + "-" + ytdl.Extractor
		}

		addToDb(el, exist)

		go deleteInteraction(s, i, c)

		server[i.GuildID].stream.Lock()
		file, _ := os.OpenFile(cachePath+el.id+audioExtension, os.O_CREATE|os.O_WRONLY, 0644)
		cmds[2].Stdout = file

		cmdsStart(cmds)
		cmdsWait(cmds)
		_ = file.Close()
		server[i.GuildID].stream.Unlock()
	} else {
		go deleteInteraction(s, i, c)
	}

	return &el, nil
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
