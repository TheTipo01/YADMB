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
	"math/rand"
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
)

func init() {

	lit.LogLevel = lit.LogError

	rand.Seed(time.Now().UnixNano())

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
		execQuery(tblSong, tblLinkLite, tblCommands)
		ignoreType = "OR"
	case mysql:
		execQuery(tblSong, tblLinkMy, tblCommands)
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
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
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

	// Checks for unused commands and deletes them
	if cmds, err := s.ApplicationCommands(s.State.User.ID, ""); err == nil {
		found := false

		for _, l := range commands {
			found = false

			for _, o := range cmds {
				// We compare every online command with the ones locally stored, to find if a command with the same name exists
				if l.Name == o.Name {
					// If the options of the command are not equal, we re-register it
					if !isCommandEqual(l, o) {
						lit.Info("Re-registering command `%s`", l.Name)

						_, err = s.ApplicationCommandCreate(s.State.User.ID, "", l)
						if err != nil {
							lit.Error("Cannot create '%s' command: %s", l.Name, err)
						}
					}

					found = true
					break
				}
			}

			// If we didn't found a match for the locally stored command, it means the command is new. We register it
			if !found {
				lit.Info("Registering new command `%s`", l.Name)

				_, err = s.ApplicationCommandCreate(s.State.User.ID, "", l)
				if err != nil {
					lit.Error("Cannot create '%s' command: %s", l.Name, err)
				}
			}
		}
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
		var err error

		server[v.GuildID].pause.Lock()

		lit.Debug("moving to " + v.ChannelID)

		_ = server[v.GuildID].vc.Disconnect()

		server[v.GuildID].vc, err = s.ChannelVoiceJoin(v.GuildID, v.ChannelID, false, true)
		if err != nil {
			// Send error and join old vc
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, cantJoinVC).
				SetColor(0x7289DA).MessageEmbed, server[v.GuildID].queue[0].txtChannel, time.Second*5)

			server[v.GuildID].vc, err = s.ChannelVoiceJoin(v.GuildID, server[v.GuildID].queue[0].channel, false, true)
		} else {
			server[v.GuildID].queue[0].channel = v.ChannelID
		}

		server[v.GuildID].pause.Unlock()

	}
}
