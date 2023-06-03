package main

import (
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/database/mysql"
	"github.com/TheTipo01/YADMB/database/sqlite"
	"github.com/TheTipo01/YADMB/spotify"
	"github.com/TheTipo01/YADMB/status"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// Holds all the info about a server
	server = make(map[string]*Server)
	// String for storing the owner of the bot
	owner string
	// spotify client
	spt *spotify.Spotify
	// Discord bot token
	token string
	// database connection
	db *database.Database
	// Cache for the blacklist
	blacklist = make(map[string]bool)
	// Discord bot session
	s *discordgo.Session
	// Holds the number of servers the bot is in
	stat status.Status
)

func init() {
	lit.LogLevel = lit.LogError

	var cfg Config
	err := fig.Load(&cfg, fig.File("config.yml"), fig.Dirs(".", "./data"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	// Config file found
	token = cfg.Token
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

	spt, err = spotify.NewSpotify(cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		lit.Error("spotify: couldn't get token: %s", err)
	}

	// Initialize the database
	switch cfg.Driver {
	case "sqlite", "sqlite3":
		db = sqlite.NewDatabase(cfg.DSN)
	case "mysql":
		db = mysql.NewDatabase(cfg.DSN)
	}

	// And load custom commands from the db
	commands, _ := db.GetCustomCommands()
	for k := range commands {
		if server[k] == nil {
			initializeServer(k)
		}

		server[k].custom = commands[k]
	}

	// Create folders used by the bot
	if _, err = os.Stat(cachePath); err != nil {
		if err = os.Mkdir(cachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", cachePath, err)
		}
	}

	// If yt-dlp is not terminated gracefully when downloading, it will leave a file called --Frag1
	_ = os.Remove("--Frag1")

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

	// Initialize the status
	stat = status.NewStatus()
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
	dg.AddHandler(guildDelete)
	dg.AddHandler(voiceStateUpdate)

	// Add commands handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Ignores commands from DM
		if i.User == nil {
			if _, ok := blacklist[i.Member.User.ID]; ok {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle,
					"User is in blacklist!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
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

	// Save the session
	s = dg

	// Register commands
	for _, v := range commands {
		_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			lit.Error("Can't register command %s: %s", v.Name, err.Error())
		}
	}

	// Wait here until CTRL-C or another term signal is received.
	lit.Info("YADMB is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
	// And the DB connection
	db.Close()
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {
	if ok, guilds := stat.CompareAndUpdate(len(s.State.Guilds)); ok {
		// Set the playing status.
		err := s.UpdateGameStatus(0, "Serving "+strconv.Itoa(guilds)+" guilds!")
		if err != nil {
			lit.Error("Can't set status, %s", err)
		}
	}
}

// Initialize Server structure
func guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.ID)

	// Populate the voiceChannelMembers map
	for _, c := range e.Channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			server[e.ID].voiceChannelMembers[c.ID] = &atomic.Int32{}
		}
	}

	// And count the members in each voice channel
	for _, v := range e.VoiceStates {
		server[e.ID].voiceChannelMembers[v.ChannelID].Add(1)
	}

	ready(s, nil)
}

func guildDelete(s *discordgo.Session, e *discordgo.GuildDelete) {
	if !server[e.ID].queue.IsEmpty() {
		clearAndExit(e.ID)
	}

	// Update the status
	ready(s, nil)
}

// Update the voice channel when the bot is moved
func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// If the bot is moved to another channel
	if v.UserID == s.State.User.ID && !server[v.GuildID].queue.IsEmpty() {
		if v.ChannelID == "" {
			// If the bot has been disconnected from the voice channel, reconnect it
			var err error

			server[v.GuildID].vc, err = s.ChannelVoiceJoin(v.GuildID, server[v.GuildID].voiceChannel, false, true)
			if err != nil {
				lit.Error("Can't join voice channel, %s", err)
			}
		} else {
			// Update the voice channel
			server[v.GuildID].voiceChannel = v.ChannelID
		}
	}

	// Update the voice channel members
	if v.ChannelID != "" {
		if v.BeforeUpdate != nil {
			server[v.GuildID].voiceChannelMembers[v.BeforeUpdate.ChannelID].Add(-1)
		}
		server[v.GuildID].voiceChannelMembers[v.ChannelID].Add(1)
	} else {
		server[v.GuildID].voiceChannelMembers[v.BeforeUpdate.ChannelID].Add(-1)
	}

	// If the bot is alone in the voice channel, stop the music
	if server[v.GuildID].voiceChannel != "" && server[v.GuildID].voiceChannelMembers[server[v.GuildID].voiceChannel].Load() == 1 {
		go quitIfEmptyVoiceChannel(v.GuildID)
	}
}
