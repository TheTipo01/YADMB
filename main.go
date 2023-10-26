package main

import (
	e "embed"
	"github.com/TheTipo01/YADMB/api"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database/mysql"
	"github.com/TheTipo01/YADMB/database/sqlite"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/TheTipo01/YADMB/spotify"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/gin-gonic/gin"
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
	server = make(map[string]*manager.Server)
	// String for storing the owner of the bot
	owner string
	// Discord bot token
	token string
	// Cache for the blacklist
	blacklist map[string]bool
	// Clients
	clients manager.Clients
	// Web API
	webApi *api.Api
	// Array of long lived tokens
	longLivedTokens []apiToken
	//go:embed all:web/build/*
	buildFS e.FS
)

func init() {
	lit.LogLevel = lit.LogError
	gin.SetMode(gin.ReleaseMode)

	var cfg Config
	err := fig.Load(&cfg, fig.File("config.yml"), fig.Dirs(".", "./data"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	// Config file found
	token = cfg.Token
	owner = cfg.Owner
	longLivedTokens = cfg.ApiTokens

	// Set lit.LogLevel to the given value
	switch strings.ToLower(cfg.LogLevel) {
	case "logwarning", "warning":
		lit.LogLevel = lit.LogWarning

	case "loginformational", "informational":
		lit.LogLevel = lit.LogInformational

	case "logdebug", "debug":
		lit.LogLevel = lit.LogDebug
	}

	if cfg.ClientID != "" && cfg.ClientSecret != "" {
		clients.Spotify, err = spotify.NewSpotify(cfg.ClientID, cfg.ClientSecret)
		if err != nil {
			lit.Error("spotify: couldn't get token: %s", err)
		}
	}

	// Start the API, if enabled
	if cfg.Address != "" {
		webApi = api.NewApi(server, cfg.Address, owner, &clients, &buildFS)
	}

	// Initialize the database
	switch cfg.Driver {
	case "sqlite", "sqlite3":
		clients.Database = sqlite.NewDatabase(cfg.DSN)
	case "mysql":
		clients.Database = mysql.NewDatabase(cfg.DSN)
	}

	// And load custom commands from the db
	commands, _ := clients.Database.GetCustomCommands()
	for k := range commands {
		if server[k] == nil {
			initializeServer(k)
		}

		server[k].Custom = commands[k]
	}

	// Load the blacklist
	blacklist, err = clients.Database.GetBlacklist()
	if err != nil {
		lit.Error("Error loading blacklist: %s", err)
	}

	// Load the DJ settings
	dj, err := clients.Database.GetDJ()
	if err != nil {
		lit.Error("Error loading DJ settings: %s", err)
	}

	for k := range dj {
		if server[k] == nil {
			initializeServer(k)
		}

		server[k].DjMode = dj[k].Enabled
		server[k].DjRole = dj[k].Role
	}

	// Create folders used by the bot
	if _, err = os.Stat(constants.CachePath); err != nil {
		if err = os.Mkdir(constants.CachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", constants.CachePath, err)
		}
	}

	// If yt-dlp is not terminated gracefully when downloading, it will leave a file called --Frag1
	_ = os.Remove("--Frag1")

	// Checks useful for knowing if every dependency
	if manager.IsCommandNotAvailable("dca") {
		lit.Error("Error: can't find dca!")
	}

	if manager.IsCommandNotAvailable("ffmpeg") {
		lit.Error("Error: can't find ffmpeg!")
	}

	if manager.IsCommandNotAvailable("yt-dlp") {
		lit.Error("Error: can't find yt-dlp!")
	}

	if cfg.YouTubeAPI != "" {
		clients.Youtube, err = youtube.NewYoutube(cfg.YouTubeAPI)
		if err != nil {
			lit.Error("youtube: couldn't get client: %s", err)
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
	dg.AddHandler(guildDelete)
	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(channelCreate)
	dg.AddHandler(guildMemberUpdate)

	// Add commands handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Ignores commands from DM
		if i.User == nil {
			if _, ok := blacklist[i.Member.User.ID]; ok {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
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
	clients.Discord = dg

	// Register commands
	_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, "", commands)
	if err != nil {
		lit.Error("Can't register commands, %s", err)
	}

	if webApi != nil {
		go webApi.HandleNotifications()

		if len(longLivedTokens) > 0 {
			lit.Info("Loading long lived tokens")
			for _, t := range longLivedTokens {
				userInfo := api.UserInfo{
					LongLivedToken: t.Token,
					Guild:          t.Guild,
					TextChannel:    t.TextChannel,
				}
				user, _ := dg.User(t.UserID)
				webApi.AddLongLivedToken(user, userInfo)
			}
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
	clients.Database.Close()
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {
	// Set the playing status.
	err := s.UpdateGameStatus(0, "Serving "+strconv.Itoa(len(s.State.Guilds))+" guilds!")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}
}

// Initialize Server structure
func guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	initializeServer(e.ID)

	// Populate the voiceChannelMembers map
	for _, c := range e.Channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			server[e.ID].VoiceChannelMembers[c.ID] = &atomic.Int32{}
		}
	}

	// And count the members in each voice channel
	for _, v := range e.VoiceStates {
		server[e.ID].VoiceChannelMembers[v.ChannelID].Add(1)
	}

	ready(s, nil)
}

func guildDelete(s *discordgo.Session, e *discordgo.GuildDelete) {
	if server[e.ID].IsPlaying() {
		manager.ClearAndExit(server[e.ID])
	}

	// Update the status
	ready(s, nil)
}

// Update the voice channel when the bot is moved
func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// If the bot is moved to another channel
	if v.UserID == s.State.User.ID && server[v.GuildID].IsPlaying() {
		if v.ChannelID == "" {
			// If the bot has been disconnected from the voice channel, reconnect it
			var err error

			lit.Debug("Reconnecting to voice channel %s on guild %s", server[v.GuildID].VoiceChannel, v.GuildID)
			server[v.GuildID].VC, err = s.ChannelVoiceJoin(v.GuildID, server[v.GuildID].VoiceChannel, false, true)
			if err != nil {
				lit.Error("Can't join voice channel, %s", err)
			}
		} else {
			// Update the voice channel
			server[v.GuildID].VoiceChannel = v.ChannelID
		}
	}

	// Update the voice channel members
	if v.ChannelID != "" {
		if v.BeforeUpdate != nil {
			server[v.GuildID].VoiceChannelMembers[v.BeforeUpdate.ChannelID].Add(-1)
		}
		server[v.GuildID].VoiceChannelMembers[v.ChannelID].Add(1)
	} else {
		server[v.GuildID].VoiceChannelMembers[v.BeforeUpdate.ChannelID].Add(-1)
	}

	// If the bot is alone in the voice channel, stop the music
	if server[v.GuildID].VoiceChannel != "" && server[v.GuildID].VoiceChannelMembers[server[v.GuildID].VoiceChannel].Load() == 1 {
		go manager.QuitIfEmptyVoiceChannel(server[v.GuildID])
	}
}

func channelCreate(_ *discordgo.Session, c *discordgo.ChannelCreate) {
	if c.Type == discordgo.ChannelTypeGuildVoice && server[c.GuildID].VoiceChannelMembers[c.ID] == nil {
		server[c.GuildID].VoiceChannelMembers[c.ID] = &atomic.Int32{}
	}
}

func guildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	// If we've been timed out, stop the music
	if m.User.ID == s.State.User.ID && m.CommunicationDisabledUntil != nil && server[m.GuildID].IsPlaying() {
		manager.ClearAndExit(server[m.GuildID])
	}
}
