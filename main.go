package main

import (
	"context"
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kkyr/fig"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Config holds data parsed from the config.yml
type Config struct {
	Token        string `fig:"token" validate:"required"`
	Owner        string `fig:"owner" validate:"required"`
	ClientID     string `fig:"clientid" validate:"required"`
	ClientSecret string `fig:"clientsecret" validate:"required"`
	DSN          string `fig:"datasourcename" validate:"required"`
	Driver       string `fig:"drivername" validate:"required"`
	Genius       string `fig:"genius" validate:"required"`
	LogLevel     string `fig:"loglevel" validate:"required"`
}

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
	// SQLite and MySQL have different syntax to ignore errors on insert
	ignoreType string
	// Cache for the blacklist
	blacklist = make(map[string]bool)
)

func init() {
	lit.LogLevel = lit.LogError

	var cfg Config
	err := fig.Load(&cfg, fig.File("config.yml"), fig.Dirs(".", "./data"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	// Check to make sure we are using one of supported drivers
	switch cfg.Driver {
	case sqlite, mysql:
	default:
		lit.Error("Wrong db driver! Valid drivers are %s and %s", sqlite, mysql)
		return
	}

	// Config file found
	token = cfg.Token
	genius = cfg.Genius
	owner = cfg.Owner

	// Set lit.LogLevel to the given value
	switch strings.ToLower(cfg.LogLevel) {
	case "logwarning", "warning":
		lit.LogLevel = lit.LogWarning

	case "loginformational", "informational":
		lit.LogLevel = lit.LogInformational

	case "logdebug", "debug":
		lit.LogLevel = lit.LogDebug
	}

	// Spotify credentials
	config := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     spotify.TokenURL,
	}

	// Check spotify token and create spotify client
	token, err := config.Token(context.Background())
	if err != nil {
		lit.Error("Spotify: couldn't get token: %s", err)
	}

	client = spotify.Authenticator{}.NewClient(token)

	// Open database connection
	db, err = sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		lit.Error("Error opening db connection, %s", err)
		return
	}

	// Create tables used by the bots
	switch cfg.Driver {
	case sqlite:
		execQuery(tblSong, tblLinkLite, tblCommands, tblBlacklist)
		ignoreType = "OR"
	case mysql:
		execQuery(tblSong, tblLinkMy, tblCommands, tblBlacklist)
		ignoreType = ""
	}

	// And load custom commands from the db
	loadCustomCommands()

	// Create folders used by the bot
	if _, err = os.Stat(cachePath); err != nil {
		if err = os.Mkdir(cachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", cachePath, err)
		}
	}

	// Checks useful for knowing if every dependency
	if isCommandNotAvailable("dca") {
		lit.Error("Error: can't find dca!")
	}

	if isCommandNotAvailable("ffmpeg") {
		lit.Error("Error: can't find ffmpeg!")
	}

	if isCommandNotAvailable("yt-dlp") {
		lit.Error("Error: can't find yt-dlp!")
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
		// Ignores commands from DM
		if i.User == nil {
			if _, ok := blacklist[i.Member.User.ID]; ok {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle,
					"User is in blacklist!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
				return
			} else {
				if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
					h(s, i)
				}
			}
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

	// Register commands
	for _, v := range commands {
		_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			lit.Error("Can't register command %s: %s", v.Name, err.Error())
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	lit.Info("YADMB is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
	// And the DB connection
	_ = db.Close()
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

// Update the voice channel when the bot is moved
func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// If the bot is moved to another channel
	if v.UserID == s.State.User.ID && v.ChannelID != "" && len(server[v.GuildID].queue) > 0 {
		// Update the voice channel
		server[v.GuildID].queue[0].channel = v.ChannelID
	}
}
