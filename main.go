package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	server = make(map[string]*sync.Mutex)
	stop   = make(map[string]bool)
	skip   = make(map[string]bool)
	queue  = make(map[string][]Queue)
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Prefix, "p", "", "Prefix for bot commands")
	flag.Parse()
}

var (
	Token  string
	Prefix string
)

func main() {

	if Token == "" {
		fmt.Println("No Token provided. Please run: discordMusicBot -t <bot Token>")
		return
	}

	if Prefix == "" {
		fmt.Println("No Prefix provided. Please run: discordMusicBot -p <prefix>")
		return
	}

	// Create a new Discord session using the provided bot Token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("discordMusicBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
}

func guildCreate(_ *discordgo.Session, event *discordgo.GuildCreate) {
	server[event.ID] = &sync.Mutex{}
	stop[event.ID] = true
	skip[event.ID] = false
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(strings.ToLower(m.Content), Prefix+"play") {
		go deleteMessage(s, m)
		link := strings.TrimPrefix(m.Content, Prefix+"play ")

		if isValidUrl(link) {
			downloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username)
		} else {
			if strings.HasPrefix(link, "spotify:playlist:") {
				spotifyPlaylist(s, m.GuildID, findUserVoiceState(s, m), m.Author.Username, strings.TrimPrefix(m.Content, Prefix+"spotify "))
			} else {
				searchDownloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username)
			}
		}
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), Prefix+"stop") {
		go deleteMessage(s, m)

		stop[m.GuildID] = false
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), Prefix+"clear") {
		go deleteMessage(s, m)

		stop[m.GuildID] = false
		skip[m.GuildID] = true
		time.Sleep(time.Millisecond * 500)
		skip[m.GuildID] = false
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), Prefix+"queue") {
		go deleteMessage(s, m)

		var message string

		//Generate song info for message
		for i, el := range queue[m.GuildID] {
			if i == 0 {
				if el.title != "" {
					message += "Currently playing: " + el.title + " - " + el.duration + " added by " + el.user + "\n\n"
					continue
				} else {
					message += "Currently playing: Getting info...\n\n"
					continue
				}

			}
			if el.title == "" {
				message += strconv.Itoa(i) + ") Getting info...\n"
			} else {
				message += strconv.Itoa(i) + ") " + el.title + " - " + el.duration + " by " + el.user + "\n"
			}

		}

		//Send embed
		em, err := s.ChannelMessageSendEmbed(m.ChannelID, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", message).SetColor(0x7289DA).MessageEmbed)
		if err != nil {
			fmt.Println("Error sending queue embed: ", err)
			return
		}

		//Wait for 15 seconds, then delete the message
		time.Sleep(time.Second * 15)
		err = s.ChannelMessageDelete(m.ChannelID, em.ID)
		if err != nil {
			fmt.Println("Error deleting queue embed: ", err)
		}

		return
	}

}
