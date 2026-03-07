package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TheTipo01/YADMB/api"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var (
	// Commands
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "play",
			Description: "Plays a song from youtube or spotify playlist (or searches the query on youtube)",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Link or query to play",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "shuffle",
					Description: "Whether to shuffle the playlist or not",
					Required:    false,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "loop",
					Description: "Whether to loop the song or not",
					Required:    false,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "playlist",
			Description: "Plays a playlist from youtube or spotify (or searches the query on youtube)",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Link or query to play",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "shuffle",
					Description: "Whether to shuffle the playlist or not",
					Required:    false,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "loop",
					Description: "Whether to loop the song or not",
					Required:    false,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "skip",
			Description: "Skips the currently playing song",
		},
		discord.SlashCommandCreate{
			Name:        "clear",
			Description: "Clears the entire queue",
		},
		discord.SlashCommandCreate{
			Name:        "pause",
			Description: "Pauses current song",
		},
		discord.SlashCommandCreate{
			Name:        "resume",
			Description: "Resumes current song",
		},
		discord.SlashCommandCreate{
			Name:        "queue",
			Description: "Returns all the songs in the server queue",
		},
		discord.SlashCommandCreate{
			Name:        "disconnect",
			Description: "Disconnect the bot from the voice channel",
		},
		discord.SlashCommandCreate{
			Name:        "restart",
			Description: "Restarts the bot (works only for the bot owner)",
		},
		discord.SlashCommandCreate{
			Name:        "addcustom",
			Description: "Creates a custom command to play a song or playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "custom-command",
					Description: "Name of the custom command",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Link to the song/playlist",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "loop",
					Description: "If you want to loop the song when called, set this to true",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "rmcustom",
			Description: "Removes a custom command",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "custom-command",
					Description: "Name of the custom command",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "stats",
			Description: "Stats™",
		},
		discord.SlashCommandCreate{
			Name:        "goto",
			Description: "Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "time",
					Description: "Time to skip to",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "listcustom",
			Description: "Lists all of the custom commands for the given server",
		},
		discord.SlashCommandCreate{
			Name:        "custom",
			Description: "Plays a song for a given custom command",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "command",
					Description: "Command to play",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "stream",
			Description: "Streams a song from a given URL",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "url",
					Description: "URL to play",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "priority",
					Description: "Does this stream have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "update",
			Description: "Update info about a song, segments from SponsorBlock, or re-download the song",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Song to update",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "info",
					Description: "Update info like thumbnail and title",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "song",
					Description: "Re-downloads the song",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "blacklist",
			Description: "Add or remove a person from the bot blacklist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "User to remove or add to the blacklist",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "dj",
			Description: "Toggles DJ mode, which allows only users with the DJ role to control the bot.",
		},
		discord.SlashCommandCreate{
			Name:        "djrole",
			Description: "Adds or removes a role from the DJ role list.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionRole{
					Name:        "role",
					Description: "Role to remove or add to the DJ role list",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "djroletoggle",
			Description: "Adds or removes the DJ role from the specified user",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "User to add or remove the DJ role from",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "webui",
			Description: "Generates a link to the web UI, which allows you to control the bot from a web browser.",
		},
	}

	// Handler
	commandHandlers = map[string]func(e *events.ApplicationCommandInteractionCreate){
		// Plays a song from YouTube or spotify playlist. If it's not a valid link, it will insert into the queue the first result for the given queue
		"play": func(e *events.ApplicationCommandInteractionCreate) {
			go server[*e.GuildID()].PlayCommand(&clients, e, false, owners)
		},
		// Plays a playlist from YouTube or spotify (or searches the query on YouTube)
		"playlist": func(e *events.ApplicationCommandInteractionCreate) {
			go server[*e.GuildID()].PlayCommand(&clients, e, true, owners)
		},
		// Skips the currently playing song
		"skip": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID) != nil {
				if server[guildID].IsPlaying() {
					el := server[guildID].Queue.GetFirstElement()
					server[guildID].Skip <- manager.Skip
					server[guildID].Paused.Store(false)

					if server[guildID].Queue.GetLength() <= 1 {
						server[guildID].ChanQuitVC <- true
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SkipTitle,
						el.Title+" - "+el.Duration+" added by "+el.User, false).
						WithColor(0x7289DA).WithThumbnail(el.Thumbnail), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SkipTitle, constants.QueueEmpty, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},

		// Clears the entire queue
		"clear": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID) != nil {
				if server[guildID].IsPlaying() {
					go server[guildID].Clean()
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.QueueTitle, constants.QueueCleared, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.QueueTitle, constants.QueueEmpty, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		"queue": func(e *events.ApplicationCommandInteractionCreate) {
			const maxQueue = 10
			guildID := *e.GuildID()

			if server[guildID].IsPlaying() {
				el := server[guildID].Queue.GetAllQueue()
				builder := discord.NewEmbed().WithTitle(manager.BotName).WithDescription(constants.QueueTitle).AddField("1", fmt.Sprintf("[%s](%s) - %s/%s added by %s\n", el[0].Title, el[0].Link,
					manager.FormatDuration(float64(server[guildID].Frames.Load())/constants.FrameSeconds), el[0].Duration, el[0].User), false)

				var nEl int
				if len(el) > maxQueue {
					nEl = maxQueue
				} else {
					nEl = len(el)
				}

				// Generate song info for the message
				for j := 1; j < nEl; j++ {
					builder = builder.AddField(strconv.Itoa(j+1), fmt.Sprintf("[%s](%s) - %s added by %s\n", el[j].Title, el[j].Link, el[j].Duration, el[j].User), false)
				}

				// Add the number of songs not shown if the queue is longer than maxQueue
				if len(el) > maxQueue {
					builder = builder.AddField("...", "And "+strconv.Itoa(len(el)-maxQueue)+" more", false)
				}

				// Send embed
				embed.SendAndDeleteEmbedInteraction(builder.WithColor(0x7289DA), e, time.Second*20, nil)
			} else {
				// Queue is empty
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.QueueTitle, constants.QueueEmpty, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		"pause": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID) != nil {
				if server[guildID].IsPlaying() {
					if server[guildID].Paused.CompareAndSwap(false, true) {
						server[guildID].Pause <- struct{}{}
						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.PauseTitle, constants.Paused, false).
							WithColor(0x7289DA), e, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.PauseTitle, constants.AlreadyPaused, false).
							WithColor(0x7289DA), e, time.Second*5, nil)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.PauseTitle, constants.QueueEmpty, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		"resume": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID) != nil {
				if server[guildID].IsPlaying() {
					if server[guildID].Paused.CompareAndSwap(true, false) {
						server[guildID].Resume <- struct{}{}
						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ResumeTitle, constants.Resumed, false).
							WithColor(0x7289DA), e, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ResumeTitle, constants.AlreadyResumed, false).
							WithColor(0x7289DA), e, time.Second*5, nil)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ResumeTitle, constants.QueueEmpty, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		"disconnect": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()
			c := embed.DeferResponse(e)

			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID) != nil {
				if !server[guildID].IsPlaying() {
					server[guildID].VC.Disconnect()
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DisconnectedTitle, constants.Disconnected, false).
						WithColor(0x7289DA), e, time.Second*5, c)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.StillPlaying, false).
						WithColor(0x7289DA), e, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, c)
			}
		},
		// Restarts the bot
		"restart": func(e *events.ApplicationCommandInteractionCreate) {
			// Check if the owner of the bot is the one who sent the command
			if _, isOwner := owners[e.Member().User.ID]; isOwner {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.RestartTitle, constants.Disconnected, false).
					WithColor(0x7289DA), e, time.Second*1, nil)

				clients.Database.Close()
				os.Exit(0)
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, "I'm sorry "+e.Member().User.Username+", I'm afraid I can't do that", false).WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		// Creates a custom command to play a song or playlist
		"addcustom": func(e *events.ApplicationCommandInteractionCreate) {
			options := e.SlashCommandInteractionData()
			command := strings.ToLower(options.String("custom-command"))
			link := options.String("link")
			loop := options.Bool("loop")
			guildID := *e.GuildID()

			if server[guildID].Custom[command] == nil {
				err := clients.Database.AddCommand(command, link, guildID, loop)
				server[guildID].Custom[command] = &database.CustomCommand{Link: link, Loop: loop}

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, err.Error(), false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SuccessfulTitle, constants.CommandAdded, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.CommandExists, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		// Removes a custom command from the DB
		"rmcustom": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if command := e.SlashCommandInteractionData().String("custom-command"); server[guildID].Custom[command] != nil {
				err := clients.Database.RemoveCustom(command, guildID)
				delete(server[guildID].Custom, command)

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, err.Error(), false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SuccessfulTitle, constants.CommandRemoved, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.CommandNotExists, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		// Lists all custom commands for the current server
		"listcustom": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			commands := make([]string, 0, len(server[guildID].Custom))

			for c := range server[guildID].Custom {
				commands = append(commands, c)
			}

			sort.Strings(commands)

			embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.CommandsTitle, strings.Join(commands, ", "), false).
				WithColor(0x7289DA), e, time.Second*30, nil)
		},
		// Calls a custom command
		"custom": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			c := embed.DeferResponse(e)

			if server[guildID].DjModeCheck(e.Member().Member, owners) {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.DjNot, false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}

			options := e.SlashCommandInteractionData()

			command := strings.ToLower(options.String("command"))

			if server[guildID].Custom[command] != nil {
				// Check if user is not in a voice channel
				if vs := manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID); vs != nil {
					if manager.JoinVC(e, *vs.ChannelID, server[guildID], c) {
						p := manager.PlayEvent{
							Username:    e.Member().User.Username,
							Song:        server[guildID].Custom[command].Link,
							Clients:     &clients,
							Event:       e,
							Random:      false,
							Loop:        server[guildID].Custom[command].Loop,
							Priority:    false,
							IsDeferred:  c,
							TextChannel: e.Channel().ID(),
						}

						if priority, ok := options.OptBool("priority"); ok {
							p.Priority = priority
						}
						server[guildID].Play(p)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
						WithColor(0x7289DA), e, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.CommandInvalid, false).
					WithColor(0x7289DA), e, time.Second*5, c)
			}
		},
		// Stats, like latency, and the size of the local cache
		"stats": func(e *events.ApplicationCommandInteractionCreate) {
			size, files := manager.FolderStats(constants.CachePath)

			embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.StatsTitle, "Called by "+e.Member().User.Username, false).
				AddField("Latency", e.Client().Gateway.Latency().String(), false).AddField("Guilds", strconv.Itoa(e.Client().Caches.GuildsLen()), false).
				AddField("Cached song", strconv.Itoa(files)+", "+
					manager.ByteCountSI(size), false).WithColor(0x7289DA), e, time.Second*15, nil)
		},
		// Refreshes things about a song
		"update": func(e *events.ApplicationCommandInteractionCreate) {
			var (
				options = e.SlashCommandInteractionData()
				link    = options.String("link")
				info    = options.Bool("info")
				song    = options.Bool("song")
			)

			if manager.IsValidURL(link) {
				if el, err := clients.Database.CheckInDb(link); err != nil {
					// Check if it's a playlist
					if entries, err := clients.Database.GetPlaylist(link); err == nil && len(entries) > 0 {
						err := clients.Database.RemovePlaylist(link)
						if err != nil {
							lit.Error("Error while removing playlist from db: %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SuccessfulTitle,
							constants.UpdateQueued, false).WithColor(0x7289DA), e, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotCached, false).
							WithColor(0x7289DA), e, time.Second*5, nil)
					}
				} else {
					if info {
						clients.Database.RemoveFromDB(el)
					}

					if song {
						err := os.Remove(constants.CachePath + el.ID + constants.AudioExtension)
						if err != nil {
							lit.Error(err.Error())
						}
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SuccessfulTitle,
						constants.UpdateQueued, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				// Check if it's in the search results
				if search, err := clients.Database.GetSearch(link); err == nil && search != "" {
					err := clients.Database.RemoveSearch(link)
					if err != nil {
						lit.Error("Error while removing search from db: %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.SuccessfulTitle,
						constants.UpdateQueued, false).WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.InvalidURL, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			}
		},
		"blacklist": func(e *events.ApplicationCommandInteractionCreate) {
			if _, isOwner := owners[e.Member().User.ID]; isOwner {
				if id := e.SlashCommandInteractionData().User("user").ID; id == e.Member().User.ID {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
						"You are really trying to add yourself to the blacklist?", false).
						WithColor(0x7289DA), e, time.Second*3, nil)
				} else {
					if _, ok := blacklist.Load(id); ok {
						// Removing from the blacklist
						blacklist.Delete(id)

						err := clients.Database.RemoveFromBlacklist(id)
						if err != nil {
							lit.Error("Error while deleting from blacklist, %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.BlacklistTitle,
							constants.BlacklistRemoved, false).
							WithColor(0x7289DA), e, time.Second*3, nil)
					} else {
						// Adding
						blacklist.Store(id, struct{}{})

						err := clients.Database.AddToBlacklist(id)
						if err != nil {
							lit.Error("Error while inserting from blacklist, %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.BlacklistTitle,
							constants.BlacklistAdded, false).
							WithColor(0x7289DA), e, time.Second*3, nil)
					}
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!", false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}
		},
		// Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...
		"goto": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if server[guildID].IsPlaying() {
				t, err := time.ParseDuration(e.SlashCommandInteractionData().String("time"))
				if err != nil {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.GotoInvalid, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					server[guildID].Queue.ModifyFirstElement(func(e *queue.Element) {
						if e.Segments == nil {
							e.Segments = make(map[int]struct{}, 2)
						}

						server[guildID].Paused.Store(true)
						server[guildID].Pause <- struct{}{}

						e.Segments[int(server[guildID].Frames.Load()+1)] = struct{}{}
						e.Segments[int(t.Seconds()*constants.FrameSeconds)] = struct{}{}

						server[guildID].Resume <- struct{}{}
						server[guildID].Paused.Store(false)
					})

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.GotoTitle, constants.SkippedTo+t.String(), false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NothingPlaying, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
		// Streams a song from the given URL, useful for radios
		"stream": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			c := embed.DeferResponse(e)
			if server[guildID].DjModeCheck(e.Member().Member, owners) {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.DjNot, false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}

			if vs := manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID); vs != nil {
				options := e.SlashCommandInteractionData()
				url := options.String("url")
				if !strings.HasPrefix(url, "file") && manager.IsValidURL(url) {
					go embed.SendEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).
						AddField(constants.EnqueuedTitle, url, false).WithColor(0x7289DA), e, c, c)

					stdout, cmds := manager.Stream(url)
					el := queue.Element{
						ID:          url,
						Title:       "Stream",
						Duration:    "",
						Link:        url,
						User:        e.Member().User.Username,
						TextChannel: e.Channel().ID(),
						BeforePlay: func() {
							manager.CmdsStart(cmds)
						},
						AfterPlay: func() {
							manager.CmdsKill(cmds)
						},
						Reader: stdout,
						Closer: stdout,
					}

					if manager.JoinVC(e, *vs.ChannelID, server[guildID], c) {
						go manager.DeleteInteraction(e.Client(), e, c)
						if priority, ok := options.OptBool("priority"); ok {
							server[guildID].AddSong(priority, el)
						} else {
							server[guildID].AddSong(false, el)
						}
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.InvalidURL, false).
						WithColor(0x7289DA), e, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, c)
			}
		},
		// Enables or disables DJ mode
		"dj": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if _, isOwner := owners[e.Member().User.ID]; isOwner {
				if server[guildID].DjMode {
					server[guildID].DjMode = false
					err := clients.Database.SetDJSettings(guildID, false)
					if err != nil {
						lit.Error("Error while disabling DJ mode, %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DjTitle, constants.DjDisabled, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					server[guildID].DjMode = true
					err := clients.Database.SetDJSettings(guildID, true)
					if err != nil {
						lit.Error("Error while enabling DJ mode, %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DjTitle, constants.DjEnabled, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!", false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}
		},
		// Sets the DJ role
		"djrole": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if _, isOwner := owners[e.Member().User.ID]; isOwner {
				role := e.SlashCommandInteractionData().Role("role")
				if role.ID != server[guildID].DjRole {
					server[guildID].DjRole = role.ID
					err := clients.Database.UpdateDJRole(guildID, role.ID)
					if err != nil {
						lit.Error("Error updating DJ role: %s", err.Error())
					}

					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DjTitle, constants.DjRoleChanged, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DjTitle, constants.DjRoleEqual, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!", false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}
		},
		// Adds or removes the DJ role from a user
		"djroletoggle": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if _, isOwner := owners[e.Member().User.ID]; isOwner {
				var err error
				var action string

				user := e.SlashCommandInteractionData().User("user")
				member, _ := e.Client().Caches.Member(*e.GuildID(), user.ID)

				if !manager.HasRole(member.RoleIDs, server[guildID].DjRole) {
					err = e.Client().Rest.AddMemberRole(*e.GuildID(), user.ID, server[guildID].DjRole)
					action = "added!"
				} else {
					err = e.Client().Rest.RemoveMemberRole(*e.GuildID(), user.ID, server[guildID].DjRole)
					action = "removed!"
				}

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
						"The bot doesn't have the necessary permission!", false).
						WithColor(0x7289DA), e, time.Second*3, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.DjTitle,
						"The role has been succefully "+action, false).
						WithColor(0x7289DA), e, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!", false).
					WithColor(0x7289DA), e, time.Second*3, nil)
			}
		},
		// Generates a link to the web UI
		"webui": func(e *events.ApplicationCommandInteractionCreate) {
			guildID := *e.GuildID()

			if vs := manager.FindUserVoiceState(e.Client(), *e.GuildID(), e.Member().User.ID); vs != nil {
				token := webApi.AddUser(&e.Member().Member, api.UserInfo{Guild: guildID, TextChannel: e.Channel().ID()})
				embed := discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.WebUITitle, fmt.Sprintf("%s/?token=%s&GuildId=%s", origin, token, guildID), false).WithColor(0x7289DA)

				// Send the response as ephemeral
				_ = e.CreateMessage(discord.NewMessageCreate().AddEmbeds(embed).AddFlags(discord.MessageFlagEphemeral))
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(manager.BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
					WithColor(0x7289DA), e, time.Second*5, nil)
			}
		},
	}
)
