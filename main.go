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
	"math/rand"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		lit.Error("Error creating Discord session: %s", err)
		return
	}

	// Add events handler
	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)
	dg.AddHandler(voiceStateUpdate)

	// Add commands handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	// Initialize intents that we use
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		lit.Error("Error opening Discord session: %s", err)
		return
	}

	// Checks for unused commands and deletes them
	if cmds, err := dg.ApplicationCommands(dg.State.User.ID, ""); err == nil {
		for _, c := range cmds {
			if commandHandlers[c.Name] == nil {
				_ = dg.ApplicationCommandDelete(dg.State.User.ID, "", c.ID)
				lit.Info("Deleted unused command %s", c.Name)
			}
		}
	}

	// And add commands used
	lit.Info("Adding used commands...")
	for _, v := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			lit.Error("Cannot create '%v' command: %v", v.Name, err)
		}
	}
	lit.Info("Commands added!")

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
	err := s.UpdateGameStatus(0, "Serving "+strconv.Itoa(len(s.State.Guilds))+" guilds!")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}

}

// Initialize Server structure
func guildCreate(_ *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.ID)
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
