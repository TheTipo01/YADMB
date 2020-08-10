package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
	"log"
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
	skip   = make(map[string]bool)
	clear  = make(map[string]bool)
	queue  = make(map[string][]Queue)
	vc     = make(map[string]*discordgo.VoiceConnection)
	client spotify.Client
	token  string
	prefix string
)

func init() {

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "!")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			fmt.Println("Config file not found! See example_config.yml")
			return
		}
	} else {
		//Config file found
		token = viper.GetString("token")
		prefix = viper.GetString("prefix")

		//Spotify credentials
		config := &clientcredentials.Config{
			ClientID:     viper.GetString("clientid"),
			ClientSecret: viper.GetString("clientsecret"),
			TokenURL:     spotify.TokenURL,
		}

		token, err := config.Token(context.Background())
		if err != nil {
			log.Fatalf("couldn't get token: %v", err)
			return
		}

		client = spotify.Authenticator{}.NewClient(token)

	}
}

func main() {

	if token == "" {
		fmt.Println("No token provided. Please modify config.yml")
		return
	}

	if prefix == "" {
		fmt.Println("No prefix provided. Please modify config.yml")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(ready)

	//Initialize intents that we use
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates)

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

func ready(s *discordgo.Session, _ *discordgo.Ready) {

	// Set the playing status.
	err := s.UpdateStatus(0, prefix+"help")
	if err != nil {
		fmt.Println("Can't set status,", err)
	}
}

//Initialize for every guild mutex and skip variable
func guildCreate(_ *discordgo.Session, event *discordgo.GuildCreate) {
	server[event.ID] = &sync.Mutex{}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if s.State.User.ID == m.Author.ID {
		return
	}

	switch strings.Split(strings.ToLower(m.Content), " ")[0] {
	//Plays a song
	case prefix + "play", prefix + "p":
		go deleteMessage(s, m)

		link := strings.TrimPrefix(m.Content, prefix+"play ")
		link = strings.TrimPrefix(link, prefix+"p ")

		if isValidUrl(link) {
			downloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID)
		} else {
			if strings.HasPrefix(link, "spotify:playlist:") {
				spotifyPlaylist(s, m.GuildID, findUserVoiceState(s, m), m.Author.Username, strings.TrimPrefix(m.Content, prefix+"spotify "), m.ChannelID)
			} else {
				searchDownloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID)
			}
		}
		break

		//Skips a song
	case prefix + "skip", prefix + "s":
		go deleteMessage(s, m)
		skip[m.GuildID] = true
		break

		//Clear the queue of the guild
	case prefix + "clear", prefix + "c":
		go deleteMessage(s, m)
		clear[m.GuildID] = true
		skip[m.GuildID] = true
		break

		//Prints out queue for the guild
	case prefix + "queue", prefix + "q":
		go deleteMessage(s, m)
		var message string
		if len(queue[m.GuildID]) > 0 {
			//Generate song info for message
			for i, el := range queue[m.GuildID] {
				if i == 0 {
					if el.title != "" {
						message += "Currently playing: " + el.title + " - " + formatDuration(time.Now().Sub(*el.time).Seconds()) + "/" + el.duration + " added by " + el.user + "\n\n"
						continue
					} else {
						message += "Currently playing: Getting info...\n\n"
						continue
					}

				}
				//If we don't have the title, we use some placeholder text
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
		} else {
			//Queue is empty
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", "Queue is empty!").SetColor(0x7289DA).MessageEmbed, m.ChannelID)
		}
		break

		//Disconnect the bot from the guild voice channel
	case prefix + "disconnect", prefix + "d":
		go deleteMessage(s, m)
		_ = vc[m.GuildID].Disconnect()
		vc[m.GuildID] = nil
		break

		//We summon the bot in the user current voice channel
	case prefix + "summon":
		go deleteMessage(s, m)
		var err error
		vc[m.GuildID], err = s.ChannelVoiceJoin(m.GuildID, findUserVoiceState(s, m), false, false)
		if err != nil {
			fmt.Println(err)
		}
		break

		//Prints out supported commands
	case prefix + "help", prefix + "h":
		go deleteMessage(s, m)
		mex, err := s.ChannelMessageSend(m.ChannelID, "Supported commands:\n```"+prefix+"play <link> - Plays a song from youtube or spotify playlist\n"+prefix+"queue - Returns all the songs in the server queue\n"+prefix+"summon - Make the bot join your voice channel\n"+prefix+"disconnect - Disconnect the bot from the voice channel\n"+prefix+"restart - Restarts the bot```")
		if err != nil {
			fmt.Println(err)
			break
		}

		time.Sleep(time.Second * 30)

		err = s.ChannelMessageDelete(m.ChannelID, mex.ID)
		if err != nil {
			fmt.Println(err)
		}
		break

		//Makes the bot exit
	case prefix + "restart", prefix + "r":
		os.Exit(0)
	}

}
