package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	//Mutex for queueing songs correctly
	server = make(map[string]*sync.Mutex)
	//Mutex for pausing/un-pausing songs
	pause = make(map[string]*sync.Mutex)
	//Need a boolean to check if song is paused, because the mutex is continuously locked and unlocked
	isPaused = make(map[string]bool)
	//Variable for skipping a single song
	skip = make(map[string]bool)
	//Variable for clearing the whole queue
	clear = make(map[string]bool)
	//The queue
	queue = make(map[string][]Queue)
	//Voice connection
	vc = make(map[string]*discordgo.VoiceConnection)
	//Custom commands, first map is for the guild id, second one is for the command, and the final string for the song
	custom = make(map[string]map[string]string)
	//Spotify client
	client spotify.Client
	//Genius key
	genius string
	//Discord bot token
	token string
	//Prefix for bot commands
	prefix         string
	dataSourceName string
	driverName     string
	db             *sql.DB
)

func init() {
	rand.Seed(time.Now().UnixNano())

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

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
		dataSourceName = viper.GetString("datasourcename")
		driverName = viper.GetString("drivername")
		genius = viper.GetString("genius")

		//Spotify credentials
		config := &clientcredentials.Config{
			ClientID:     viper.GetString("clientid"),
			ClientSecret: viper.GetString("clientsecret"),
			TokenURL:     spotify.TokenURL,
		}

		token, err := config.Token(context.Background())
		if err != nil {
			log.Printf("Spotify: couldn't get token: %v", err)
		}

		client = spotify.Authenticator{}.NewClient(token)

		//Database
		db, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Println("Error opening db connection,", err)
			return
		}

		execQuery(tblSong, db)
		execQuery(tblCommands, db)

		loadCustomCommands(db)

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
	pause[event.ID] = &sync.Mutex{}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
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
			downloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID, false)
		} else {
			if strings.HasPrefix(link, "spotify:playlist:") {
				spotifyPlaylist(s, m.GuildID, findUserVoiceState(s, m), m.Author.Username, strings.TrimPrefix(m.Content, prefix+"spotify "), m.ChannelID, false)
			} else {
				searchDownloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID, false)
			}
		}
		break

		//Randomly plays a song (or a playlist)
	case prefix + "shuffle":
		go deleteMessage(s, m)

		link := strings.TrimPrefix(m.Content, prefix+"shuffle ")

		if isValidUrl(link) {
			downloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID, true)
		} else {
			if strings.HasPrefix(link, "spotify:playlist:") {
				spotifyPlaylist(s, m.GuildID, findUserVoiceState(s, m), m.Author.Username, strings.TrimPrefix(m.Content, prefix+"spotify "), m.ChannelID, true)
			} else {
				searchDownloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), link, m.Author.Username, m.ChannelID, true)
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
					if el.title != "" && el.time != nil {
						if isPaused[m.GuildID] {
							message += "Currently playing: " + el.title + " - " + el.lastTime + "/" + el.duration + " added by " + el.user + "\n\n"
							continue
						} else {
							//TODO: Fix offset...
							message += "Currently playing: " + el.title + " - " + formatDuration(time.Now().Sub(*el.time).Seconds()+el.offset) + "/" + el.duration + " added by " + el.user + "\n\n"
							continue
						}
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

		message := "Supported commands:\n```" +
			prefix + "play <link> - Plays a song from youtube or spotify playlist\n" +
			prefix + "shuffle <playlist> - Shuffles the songs in the playlist and adds them to the queue" +
			prefix + "pause - Pauses current song\n" +
			prefix + "resume - Resumes current song\n" +
			prefix + "queue - Returns all the songs in the server queue\n" +
			prefix + "lyrics <song> - Tries to search for lyrics of the specified song, or if not specified searches for the title of the currently playing song\n" +
			prefix + "summon - Make the bot join your voice channel\n" +
			prefix + "disconnect - Disconnect the bot from the voice channel\n" +
			prefix + "restart - Restarts the bot\n" +
			prefix + "custom <custom_command> <song/playlist> - Creates a custom command to play a song or playlist\n" +
			prefix + "rmcustom <costom_command> - Removes a custom command\n" +
			"```"
		//If we have custom commands, we add them to the help message
		if len(custom[m.GuildID]) > 0 {
			message += "\nCustom commands:\n```"

			for k := range custom[m.GuildID] {
				message += k + ", "
			}

			message = strings.TrimSuffix(message, ", ")
			message += "```"
		}

		mex, err := s.ChannelMessageSend(m.ChannelID, message)
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

		//Pause the song
	case prefix + "pause":
		go deleteMessage(s, m)
		if !isPaused[m.GuildID] && len(queue[m.GuildID]) > 0 {
			isPaused[m.GuildID] = true
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Paused the current song").SetColor(0x7289DA).MessageEmbed, m.ChannelID)
			pause[m.GuildID].Lock()

			queue[m.GuildID][0].lastTime = formatDuration(time.Now().Sub(*queue[m.GuildID][0].time).Seconds() + queue[m.GuildID][0].offset)

			//Covering edge case where voiceConnection is not established
			if vc[m.GuildID] != nil {
				_ = vc[m.GuildID].Speaking(false)
			}
		}
		break

		//Resume playing
	case prefix + "resume":
		go deleteMessage(s, m)
		if isPaused[m.GuildID] {
			isPaused[m.GuildID] = false
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Resumed the current song").SetColor(0x7289DA).MessageEmbed, m.ChannelID)
			queue[m.GuildID][0].offset += queue[m.GuildID][0].time.Sub(time.Now()).Seconds()

			pause[m.GuildID].Unlock()
			_ = vc[m.GuildID].Speaking(true)
		}
		break

		//Adds a custom command
	case prefix + "custom":
		go deleteMessage(s, m)

		splitted := strings.Split(strings.TrimPrefix(m.Content, prefix+"custom "), " ")

		if splitted[0] != "" && splitted[1] != "" {
			addCommand(strings.ToLower(splitted[0]), splitted[1], m.GuildID)
		}
		break

		//Removes a custom command
	case prefix + "rmcustom":
		go deleteMessage(s, m)

		removeCustom(strings.TrimPrefix(m.Content, prefix+"rmcustom "), m.GuildID)
		break

	case prefix + "lyrics":
		go deleteMessage(s, m)
		song := strings.TrimPrefix(strings.TrimPrefix(m.Content, prefix+"lyrics"), " ")

		if song == "" {
			song = queue[m.GuildID][0].title
		}

		if len(queue[m.GuildID]) > 0 {

			text := formatLongMessage(lyrics(song))

			mex, _ := s.ChannelMessageSend(m.ChannelID, "Lyrics for "+song+": ")
			queue[m.GuildID][0].messageID = append(queue[m.GuildID][0].messageID, *mex)

			//If the messages are more then 3, we don't send anything
			if len(text) > 3 {
				mex, _ := s.ChannelMessageSend(m.ChannelID, "```Lyrics too long!```")
				queue[m.GuildID][0].messageID = append(queue[m.GuildID][0].messageID, *mex)
				return
			}

			for _, t := range text {
				mex, _ = s.ChannelMessageSend(m.ChannelID, "```"+t+"```")
				queue[m.GuildID][0].messageID = append(queue[m.GuildID][0].messageID, *mex)
			}

		}

		break

		//Makes the bot exit
	case prefix + "restart", prefix + "r":
		go deleteMessage(s, m)
		os.Exit(0)

		//We search for possible custom commands
	default:
		lower := strings.TrimPrefix(strings.ToLower(m.Content), prefix)

		if custom[m.GuildID][lower] != "" {
			go deleteMessage(s, m)

			if isValidUrl(custom[m.GuildID][lower]) {
				downloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), custom[m.GuildID][lower], m.Author.Username, m.ChannelID, false)
			} else {
				if strings.HasPrefix(custom[m.GuildID][lower], "spotify:playlist:") {
					spotifyPlaylist(s, m.GuildID, findUserVoiceState(s, m), m.Author.Username, strings.TrimPrefix(m.Content, prefix+"spotify "), m.ChannelID, false)
				} else {
					searchDownloadAndPlay(s, m.GuildID, findUserVoiceState(s, m), custom[m.GuildID][lower], m.Author.Username, m.ChannelID, false)
				}
			}
			break
		}

	}

}
