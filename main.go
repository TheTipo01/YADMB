package main

import (
	"context"
	e "embed"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/TheTipo01/YADMB/api"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database/mysql"
	"github.com/TheTipo01/YADMB/database/sqlite"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/TheTipo01/YADMB/spotify"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
	"github.com/kkyr/fig"
)

var (
	// Holds all the info about a server
	server = make(map[snowflake.ID]*manager.Server)
	// String for storing the owners of the bot
	owners map[snowflake.ID]struct{}
	// Discord bot token
	token string
	// Cache for the user blacklist
	blacklist *sync.Map
	// Clients
	clients manager.Clients
	// Web API
	webApi *api.Api
	// Array of long lived tokens
	longLivedTokens []apiToken
	//go:embed all:web/build/*
	buildFS e.FS
	// Origin for CORS and link generation
	origin string
	// Server mutex
	serverMutex sync.RWMutex
	// If set to true, the bot will only respond to commands coming from guilds in the guild list
	whitelist bool
	// List of guilds the bot will respond to
	guildList *sync.Map
	// Channel used to notify the presence updater that the guild count has changed
	guildCountChan = make(chan struct{})
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

	owners = make(map[snowflake.ID]struct{}, len(cfg.Owner))
	for _, o := range cfg.Owner {
		owners[o] = struct{}{}
	}

	longLivedTokens = cfg.ApiTokens
	origin = cfg.Origin

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
		webApi = api.NewApi(server, cfg.Address, owners, &clients, &buildFS, origin)
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

	// Load the whitelist
	whitelist = cfg.WhiteList
	guildList = &sync.Map{}
	for _, g := range cfg.GuildList {
		guildList.Store(g, struct{}{})
	}

	// Create folders used by the bot
	if _, err = os.Stat(constants.CachePath); err != nil {
		if err = os.Mkdir(constants.CachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", constants.CachePath, err)
		}
	}

	// If yt-dlp is not terminated gracefully when downloading, it will leave a file called --Frag1
	_ = os.Remove("--Frag1")

	// Checks useful for knowing if every dependency exists
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

	go presenceUpdater()
}

func main() {
	if token == "" {
		lit.Error("No token provided. Please modify config.yml")
		return
	}

	logger := slog.Default()
	if lit.LogLevel == lit.LogDebug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	client, _ := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuilds,
			),
		),

		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.FlagVoiceStates,
			),
		),

		bot.WithEventListenerFunc(ready),
		bot.WithEventListenerFunc(guildCreate),
		bot.WithEventListenerFunc(guildDelete),
		//bot.WithEventListenerFunc(voiceStateUpdate),
		bot.WithEventListenerFunc(guildMemberUpdate),
		bot.WithEventListenerFunc(interactionCreate),

		bot.WithVoiceManagerConfigOpts(voice.WithDaveSessionCreateFunc(golibdave.NewSession)),

		bot.WithLogger(logger),
	)

	defer client.Close(context.TODO())

	if err := client.OpenGateway(context.TODO()); err != nil {
		lit.Error("errors while connecting to gateway %s", err)
		return
	}

	// Register commands
	_, err := client.Rest.SetGlobalCommands(client.ApplicationID, commands)
	if err != nil {
		lit.Error("Error registering commands: %s", err)
		return
	}

	// Start the web API, if enabled
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
				user, err := client.Rest.GetMember(t.Guild, t.UserID)
				if err != nil {
					lit.Error("Error loading long lived token for user %s in guild %s: %s", t.UserID, t.Guild, err)
				} else {
					webApi.AddLongLivedToken(user, userInfo)
				}
			}
		}
	}

	// Print guilds the bot is connected to
	if lit.LogLevel == lit.LogDebug {
		lit.Debug("Bot is connected to %d guilds.", len(server))

		for id := range server {
			lit.Debug("Guild ID: %s", id.String())
		}

	}

	// Save the session
	clients.Discord = client

	// Wait here until CTRL-C or another term signal is received.
	lit.Info("YADMB is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// And the DB connection
	clients.Database.Close()
}

// presenceUpdater updates the bot presence every time the guild count changes, with a debounce of 500ms to avoid making too many requests
func presenceUpdater() {
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()

	for {
		select {
		case <-guildCountChan:
			debounceTimer.Reset(500 * time.Millisecond)
		case <-debounceTimer.C:
			if clients.Discord != nil {
				_ = clients.Discord.SetPresence(context.TODO(), gateway.WithCustomActivity("Serving "+strconv.Itoa(len(server))+" guilds!"))
			}
		}
	}
}

func notifyGuildCountChange() {
	select {
	case guildCountChan <- struct{}{}:
	default:
	}
}

func ready(e *events.Ready) {
	notifyGuildCountChange()

	manager.BotName = e.User.Username
}

// Initialize Server structure
func guildCreate(e *events.GuildReady) {
	initializeServer(e.GuildID)

	notifyGuildCountChange()
}

func guildDelete(e *events.GuildLeave) {
	if server[e.GuildID].IsPlaying() {
		ClearAndExit(server[e.GuildID])
	}

	// Update the status
	notifyGuildCountChange()
}

// Update the voice channel when the bot is moved
func voiceStateUpdate(v *events.GuildVoiceStateUpdate) {
	// If the bot is alone in the voice channel, stop the music
	if server[v.VoiceState.GuildID].VC.IsConnected() {
		channel := server[v.VoiceState.GuildID].VC.GetChannelID()
		if (v.VoiceState.ChannelID == channel || (v.OldVoiceState.ChannelID != nil && v.OldVoiceState.ChannelID == channel)) && countVoiceStates(v.Client(), v.VoiceState.GuildID, *channel) == 0 {
			go QuitIfEmptyVoiceChannel(server[v.VoiceState.GuildID])
		}
	}
}

func guildMemberUpdate(m *events.GuildMemberUpdate) {
	// If we've been timed out, stop the music
	if m.Member.User.ID == m.Client().ApplicationID && m.Member.CommunicationDisabledUntil != nil &&
		m.Member.CommunicationDisabledUntil.After(time.Now()) && server[m.GuildID].IsPlaying() {
		ClearAndExit(server[m.GuildID])
	}
}

func interactionCreate(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	// Ignores commands from DM
	if e.Context() == discord.InteractionContextTypeGuild {
		if _, ok := blacklist.Load(e.User().ID.String()); ok {
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(manager.BotName).AddField(constants.ErrorTitle,
				constants.UserInBlacklist, false).
				SetColor(0x7289DA).Build(), e, time.Second*3, nil)
		} else {
			if whitelist {
				// Whitelist mode: check if the guild is in the list
				if _, ok = guildList.Load(e.GuildID().String()); ok {
					if h, ok := commandHandlers[data.CommandName()]; ok {
						go h(e)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(manager.BotName).AddField(constants.ErrorTitle,
						constants.ServerNotInWhitelist, false).
						SetColor(0x7289DA).Build(), e, time.Second*3, nil)
				}
			} else {
				// Blacklist mode: check if the guild is not in the list
				if _, ok = guildList.Load(e.GuildID().String()); !ok {
					if h, ok := commandHandlers[data.CommandName()]; ok {
						go h(e)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(manager.BotName).AddField(constants.ErrorTitle,
						constants.ServerInBlacklist, false).
						SetColor(0x7289DA).Build(), e, time.Second*3, nil)
				}
			}
		}
	} else {
		if _, ok := blacklist.Load(e.User().ID.String()); ok {
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(manager.BotName).AddField(constants.ErrorTitle,
				constants.UserInBlacklist, false).
				SetColor(0x7289DA).Build(), e, time.Second*3, nil)
		} else {
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(manager.BotName).AddField(constants.ErrorTitle,
				constants.ErrorDM, false).
				SetColor(0x7289DA).Build(), e, time.Second*15, nil)
		}
	}
}
