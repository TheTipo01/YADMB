package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

//Structure for holding infos about a song
type Queue struct {
	title    string
	duration string
	id       string
	link     string
	user     string
	message  string
}

//Logs and instantly delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.Username + ": " + m.Content)
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		fmt.Println("Can't delete message,", err)
	}
}

//Finds user current voice channel
func findUserVoiceState(session *discordgo.Session, m *discordgo.MessageCreate) string {
	user := m.Author.ID

	//TODO: Better webhook handling
	//My user id, for playing song via a webhook
	if m.WebhookID != "" {
		user = "145618075452964864"
	}

	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == user {
				return vs.ChannelID
			}
		}
	}

	return ""
}

//Checks if a string is a valid URL
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

//Gets title and duration for a given video
func addInfo(id string, guild string) {
	for i, el := range queue[guild] {
		if el.id == id {
			out, _ := exec.Command("youtube-dl", "-e", "--get-duration", el.link).Output()
			output := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

			if len(output) == 2 {
				queue[guild][i].title = output[0]
				queue[guild][i].duration = output[1]
			} else {
				removeFromQueue(id, guild)
			}
			return
		}
	}
}

//Removes element from the queue
func removeFromQueue(id string, guild string) {
	for i, q := range queue[guild] {
		if q.id == id {
			copy(queue[guild][i:], queue[guild][i+1:])
			queue[guild][len(queue[guild])-1] = Queue{"", "", "", "", "", ""}
			queue[guild] = queue[guild][:len(queue[guild])-1]
			return
		}
	}
}

//Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 3)

	err = s.ChannelMessageDelete(txtChannel, m.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
}

//Finds pointer for a given song id
func findQueuePointer(guildId, id string) int {
	for i := range queue[guildId] {
		if queue[guildId][i].id == id {
			return i
		}
	}

	return -1
}
