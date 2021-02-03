package main

import (
	"context"
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"math/rand"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// How many DCA frames are needed for a second. It's not perfect, but good enough.
	frameSeconds = 50
)

var (
	// Holds all the info about a server
	server = make(map[string]*Server)
	// String for storing the owner of the bot
	owner string
	// Spotify client
	client spotify.Client
	// Genius key
	genius string
	// Discord bot token
	token string
	// Prefix for bot commands
	prefix string
	// Database connection
	db *sql.DB
)

func init() {

	lit.LogLevel = lit.LogError

	rand.Seed(time.Now().UnixNano())

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			lit.Error("Config file not found! See example_config.yml")
			return
		}
	} else {
		// Config file found
		token = viper.GetString("token")
		prefix = viper.GetString("prefix")
		genius = viper.GetString("genius")
		owner = viper.GetString("owner")

		// Set lit.LogLevel to the given value
		switch strings.ToLower(viper.GetString("loglevel")) {
		case "logerror", "error":
			lit.LogLevel = lit.LogError
			break
		case "logwarning", "warning":
			lit.LogLevel = lit.LogWarning
			break
		case "loginformational", "informational":
			lit.LogLevel = lit.LogInformational
			break
		case "logdebug", "debug":
			lit.LogLevel = lit.LogDebug
			break
		}

		// Spotify credentials
		config := &clientcredentials.Config{
			ClientID:     viper.GetString("clientid"),
			ClientSecret: viper.GetString("clientsecret"),
			TokenURL:     spotify.TokenURL,
		}

		// Check spotify token and create spotify client
		token, err := config.Token(context.Background())
		if err != nil {
			lit.Error("Spotify: couldn't get token: %s", err)
		}

		client = spotify.Authenticator{}.NewClient(token)

		// Open database connection
		db, err = sql.Open(viper.GetString("drivername"), viper.GetString("datasourcename"))
		if err != nil {
			lit.Error("Error opening db connection, %s", err)
			return
		}

		// Create tables used by the bots
		execQuery(tblSong)
		execQuery(tblCommands)

		// And load custom commands from the db
		loadCustomCommands()

		// Create folders used by the bot
		if _, err = os.Stat("./audio_cache"); err != nil {
			if err = os.Mkdir("./audio_cache", 0755); err != nil {
				lit.Error("Cannot create audio_cache directory, %s", err)
			}
		}

	}

}

func main() {

	if token == "" {
		lit.Error("No token provided. Please modify config.yml")
		return
	}

	if prefix == "" {
		lit.Error("No prefix provided. Please modify config.yml")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		lit.Error("Error creating Discord session: %s", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	dg.AddHandler(messageCreate)
	dg.AddHandler(voiceStateUpdate)

	// Initialize intents that we use
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		lit.Error("Error opening Discord session: %s", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	lit.Info("YADMB is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()

}

func ready(s *discordgo.Session, _ *discordgo.Ready) {

	// Set the playing status.
	err := s.UpdateGameStatus(0, prefix+"help")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}

}

// Initialize Server structure
func guildCreate(_ *discordgo.Session, e *discordgo.GuildCreate) {

	if server[e.ID] == nil {
		server[e.ID] = &Server{server: &sync.Mutex{}, pause: &sync.Mutex{}, stream: &sync.Mutex{}, custom: make(map[string]string)}
	}

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages sent from the bot, messages if the user is a bot, and messages without the prefix
	if s.State.User.ID == m.Author.ID || m.Author.Bot || !strings.HasPrefix(m.Content, prefix) {
		return
	}

	// Split the message on spaces
	splittedMessage := strings.Split(m.Content, " ")

	command := strings.TrimPrefix(strings.ToLower(splittedMessage[0]), prefix)

	switch command {
	// Plays a song
	case "play", "p":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m)

		// Check if user is not in a voice channel
		if vs == nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		// Check if the user also sent a song
		if len(splittedMessage) < 2 {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No song specified!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		play(s, strings.TrimPrefix(m.Content, splittedMessage[0]+" "), m.ChannelID, vs.ChannelID, m.GuildID, m.Author.Username, false)
		break

		// Randomly plays a song (or a playlist)
	case "shuffle":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m)

		// Check if user is not in a voice channel
		if vs == nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		// Check if the user also sent a song
		if len(splittedMessage) < 2 {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No song specified!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		play(s, strings.TrimPrefix(m.Content, splittedMessage[0]+" "), m.ChannelID, vs.ChannelID, m.GuildID, m.Author.Username, true)
		break

		// Skips a song
	case "skip", "s":
		go deleteMessage(s, m)

		// Check if user is not in a voice channel
		if findUserVoiceState(s, m) == nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		server[m.GuildID].skip = true
		break

		// Clear the queue of the guild
	case "clear", "c":
		go deleteMessage(s, m)

		// Check if user is not in a voice channel
		if findUserVoiceState(s, m) == nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		server[m.GuildID].clear = true
		server[m.GuildID].skip = true
		break

		// Prints out queue for the guild
	case "queue", "q":
		go deleteMessage(s, m)
		var message string

		if len(server[m.GuildID].queue) > 0 {
			// Generate song info for message
			for i, el := range server[m.GuildID].queue {
				if i == 0 {
					if el.title != "" {
						message += "Currently playing: " + el.title + " - " + formatDuration(float64(server[m.GuildID].queue[0].frame/frameSeconds)) +
							"/" + el.duration + " added by " + el.user + "\n\n"
						continue
					} else {
						message += "Currently playing: Getting info...\n\n"
						continue
					}

				}
				// If we don't have the title, we use some placeholder text
				if el.title == "" {
					message += strconv.Itoa(i) + ") Getting info...\n"
				} else {
					message += strconv.Itoa(i) + ") " + el.title + " - " + el.duration + " by " + el.user + "\n"
				}

			}

			// Send embed
			em, err := s.ChannelMessageSendEmbed(m.ChannelID, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", message).
				SetColor(0x7289DA).MessageEmbed)
			if err != nil {
				lit.Error("Error sending queue embed: %s", err)
				return
			}

			// Wait for 15 seconds, then delete the message
			time.Sleep(time.Second * 15)
			err = s.ChannelMessageDelete(m.ChannelID, em.ID)
			if err != nil {
				lit.Error("Error deleting queue embed: %s", err)
			}
		} else {
			// Queue is empty
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", "Queue is empty!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
		}
		break

		// Disconnect the bot from the guild voice channel
	case "disconnect", "d":
		go deleteMessage(s, m)

		// Check if the queue is empty
		if len(server[m.GuildID].queue) == 0 {
			server[m.GuildID].server.Lock()

			_ = server[m.GuildID].vc.Disconnect()
			server[m.GuildID].vc = nil

			server[m.GuildID].server.Unlock()
		} else {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't disconnect the bot!\nStill playing in a voice channel.").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
		}
		break

		// We summon the bot in the user current voice channel
	case "summon":
		go deleteMessage(s, m)

		vs := findUserVoiceState(s, m)

		// Check if user is not in a voice channel
		if vs == nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		var err error

		// If we are playing something, we lock the pause mutex
		if len(server[m.GuildID].queue) > 0 {
			server[m.GuildID].pause.Lock()

			// Disconnect the bot
			if server[m.GuildID].vc != nil {
				_ = server[m.GuildID].vc.Disconnect()
			}

			// And reconnect the bot to the new voice channel
			server[m.GuildID].queue[0].channel = vs.ChannelID
			server[m.GuildID].vc, err = s.ChannelVoiceJoin(m.GuildID, vs.ChannelID, false, true)

			server[m.GuildID].pause.Unlock()
		} else {
			// Else we just join the channel and wait
			server[m.GuildID].server.Unlock()

			server[m.GuildID].vc, err = s.ChannelVoiceJoin(m.GuildID, vs.ChannelID, false, true)

			// We also start the quitVC routine to disconnect the bot after a minute of inactivity
			go quitVC(m.GuildID)

			server[m.GuildID].server.Unlock()
		}

		if err != nil {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't join voice channel!\n"+err.Error()).
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break
		}

		break

		// Prints out supported commands
	case "help", "h":
		go deleteMessage(s, m)

		message := "Supported commands:\n```" +
			prefix + "play <link> - Plays a song from youtube or spotify playlist (or search the query on youtube)\n" +
			prefix + "skip - Skips the currently playing song\n" +
			prefix + "clear - Clears the entire queue\n" +
			prefix + "shuffle <playlist> - Shuffles the songs in the playlist and adds them to the queue\n" +
			prefix + "pause - Pauses current song\n" +
			prefix + "resume - Resumes current song\n" +
			prefix + "queue - Returns all the songs in the server queue\n" +
			prefix + "lyrics <song> - Tries to search for lyrics of the specified song, or if not specified searches for the title of the currently playing song\n" +
			prefix + "summon - Make the bot join your voice channel\n" +
			prefix + "disconnect - Disconnect the bot from the voice channel\n" +
			prefix + "restart - Restarts the bot (works only for the bot owner)\n" +
			prefix + "custom <custom_command> <song/playlist> - Creates a custom command to play a song or playlist\n" +
			prefix + "rmcustom <custom_command> - Removes a custom command\n" +
			prefix + "stats - Stats™\n" +
			prefix + "goto <time> - Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...\n" +
			"```"
		// If we have custom commands, we add them to the help message
		if len(server[m.GuildID].custom) > 0 {
			message += "\nCustom commands:\n```"

			for k := range server[m.GuildID].custom {
				message += k + ", "
			}

			message = strings.TrimSuffix(message, ", ")
			message += "```"
		}

		mex, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			lit.Error("MessageSend failed: %s", err)
			break
		}

		time.Sleep(time.Second * 30)

		err = s.ChannelMessageDelete(m.ChannelID, mex.ID)
		if err != nil {
			lit.Error("MessageDelete failed: %s", err)
		}
		break

		// Pause the song
	case "pause":
		go deleteMessage(s, m)

		if len(server[m.GuildID].queue) > 0 && !server[m.GuildID].isPaused {
			server[m.GuildID].isPaused = true
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Paused the current song").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			server[m.GuildID].pause.Lock()
		}
		break

		// Resume playing
	case "resume":
		go deleteMessage(s, m)

		if server[m.GuildID].isPaused {
			server[m.GuildID].isPaused = false
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Resumed the current song").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)

			server[m.GuildID].pause.Unlock()
			err := server[m.GuildID].vc.Speaking(true)
			if err != nil {
				lit.Error("vc.Speaking(true) failed: %s", err)
			}
		}
		break

		// Adds a custom command
	case "custom":
		go deleteMessage(s, m)

		switch len(splittedMessage) {
		case 0, 1, 2:
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Too few arguments!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			break

		case 3:
			err := addCommand(strings.ToLower(splittedMessage[1]), splittedMessage[2], m.GuildID)
			if err != nil {
				sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", err.Error()).
					SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			} else {
				sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Successful", "Custom command added!").
					SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			}
			break

		default:
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Too many arguments!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
		}

		break

		// Removes a custom command
	case "rmcustom":
		go deleteMessage(s, m)

		removeCustom(strings.TrimPrefix(m.Content, splittedMessage[0]+" "), m.GuildID)
		break

		// Prints lyrics of a song
	case "lyrics":
		go deleteMessage(s, m)

		// We search for lyrics only if there's something playing
		if len(server[m.GuildID].queue) > 0 {
			song := strings.TrimPrefix(m.Content, splittedMessage[0]+" ")

			// If the user didn't input a title, we use the currently playing video
			if song == "" {
				song = server[m.GuildID].queue[0].title
			}

			text := formatLongMessage(lyrics(song))

			mex, err := s.ChannelMessageSend(m.ChannelID, "Lyrics for "+song+": ")
			if err != nil {
				lit.Error("Lyrics MessageSend failed: %s", err)
				break
			}

			server[m.GuildID].queue[0].messageID = append(server[m.GuildID].queue[0].messageID, *mex)

			// If the messages are more then 3, we don't send anything
			if len(text) > 3 {
				mex, _ := s.ChannelMessageSend(m.ChannelID, "```Lyrics too long!```")
				server[m.GuildID].queue[0].messageID = append(server[m.GuildID].queue[0].messageID, *mex)
				break
			}

			for _, t := range text {
				mex, _ = s.ChannelMessageSend(m.ChannelID, "```"+t+"```")
				server[m.GuildID].queue[0].messageID = append(server[m.GuildID].queue[0].messageID, *mex)
			}

		}
		break

		// Makes the bot exit
	case "restart", "r":
		deleteMessage(s, m)

		// Check if the owner of the bot required the restart
		if owner == m.Author.ID {
			os.Exit(0)
		} else {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "I'm sorry "+m.Author.Username+", I'm afraid I can't do that").SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
		}
		break

	// Stats™
	case "stats":
		go deleteMessage(s, m)

		files, _ := ioutil.ReadDir("./audio_cache")

		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Stats™", "Called by "+m.Author.Username).
			AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
			AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(len(files))+", "+
			ByteCountSI(DirSize("./audio_cache"))).SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*15)
		break

	// Skips to a given time
	case "goto":
		go deleteMessage(s, m)

		if len(server[m.GuildID].queue) > 0 {
			t, err := time.ParseDuration(strings.TrimPrefix(m.Content, splittedMessage[0]+" "))
			if err != nil {
				sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Wrong format.\nValid formats are: 1h10m3s, 3m, 4m10s...").
					SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
			} else {
				if server[m.GuildID].queue[0].segments == nil {
					server[m.GuildID].queue[0].segments = make(map[int]bool)
				}

				server[m.GuildID].pause.Lock()

				server[m.GuildID].queue[0].segments[server[m.GuildID].queue[0].frame+1] = true
				server[m.GuildID].queue[0].segments[int(t.Seconds()*frameSeconds)] = true

				server[m.GuildID].pause.Unlock()
			}
		} else {
			sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No songs playing!").
				SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
		}
		break

		// We search for possible custom commands
	default:

		if server[m.GuildID].custom[command] != "" {
			go deleteMessage(s, m)

			vs := findUserVoiceState(s, m)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, m.ChannelID, time.Second*5)
				break
			}

			play(s, server[m.GuildID].custom[command], m.ChannelID, vs.ChannelID, m.GuildID, m.Author.Username, false)
			break
		}
		break

	}

}

// Used to reconnect the bot to the channel where it's moved
// Still a bit broken, as it first reconnect to the old voice channel, disconnect, and connect to the new channel
func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if v.UserID == s.State.User.ID && server[v.GuildID].vc != nil && len(server[v.GuildID].queue) > 0 && v.ChannelID != server[v.GuildID].queue[0].channel && v.ChannelID != "" {
		server[v.GuildID].pause.Lock()

		lit.Debug("moving to " + v.ChannelID)

		server[v.GuildID].queue[0].channel = v.ChannelID
		_ = server[v.GuildID].vc.Disconnect()

		server[v.GuildID].vc, _ = s.ChannelVoiceJoin(v.GuildID, v.ChannelID, false, true)

		server[v.GuildID].pause.Unlock()
	}
}
