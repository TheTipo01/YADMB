package main

import (
	"fmt"
	"github.com/TheTipo01/YADMB/api"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// Commands
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Plays a song from youtube or spotify playlist (or searches the query on youtube)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link or query to play",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "shuffle",
					Description: "Whether to shuffle the playlist or not",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "loop",
					Description: "Whether to loop the song or not",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		{
			Name:        "playlist",
			Description: "Plays a playlist from youtube or spotify (or searches the query on youtube)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link or query to play",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "shuffle",
					Description: "Whether to shuffle the playlist or not",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "loop",
					Description: "Whether to loop the song or not",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		{
			Name:        "skip",
			Description: "Skips the currently playing song",
		},
		{
			Name:        "clear",
			Description: "Clears the entire queue",
		},
		{
			Name:        "pause",
			Description: "Pauses current song",
		},
		{
			Name:        "resume",
			Description: "Resumes current song",
		},
		{
			Name:        "queue",
			Description: "Returns all the songs in the server queue",
		},
		{
			Name:        "disconnect",
			Description: "Disconnect the bot from the voice channel",
		},
		{
			Name:        "restart",
			Description: "Restarts the bot (works only for the bot owner)",
		},
		{
			Name:        "addcustom",
			Description: "Creates a custom command to play a song or playlist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "custom-command",
					Description: "Name of the custom command",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link to the song/playlist",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "loop",
					Description: "If you want to loop the song when called, set this to true",
					Required:    true,
				},
			},
		},
		{
			Name:        "rmcustom",
			Description: "Removes a custom command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "custom-command",
					Description: "Name of the custom command",
					Required:    true,
				},
			},
		},
		{
			Name:        "stats",
			Description: "Stats™",
		},
		{
			Name:        "goto",
			Description: "Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "time",
					Description: "Time to skip to",
					Required:    true,
				},
			},
		},
		{
			Name:        "listcustom",
			Description: "Lists all of the custom commands for the given server",
		},
		{
			Name:        "custom",
			Description: "Plays a song for a given custom command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Command to play",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "priority",
					Description: "Does this song have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		{
			Name:        "stream",
			Description: "Streams a song from a given URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "URL to play",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "priority",
					Description: "Does this stream have priority over the other songs in the queue?",
					Required:    false,
				},
			},
		},
		{
			Name:        "update",
			Description: "Update info about a song, segments from SponsorBlock, or re-download the song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Song to update",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "info",
					Description: "Update info like thumbnail and title",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "song",
					Description: "Re-downloads the song",
					Required:    true,
				},
			},
		},
		{
			Name:        "blacklist",
			Description: "Add or remove a person from the bot blacklist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to remove or add to the blacklist",
					Required:    true,
				},
			},
		},
		{
			Name:        "dj",
			Description: "Toggles DJ mode, which allows only users with the DJ role to control the bot.",
		},
		{
			Name:        "djrole",
			Description: "Adds or removes a role from the DJ role list.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "Role to remove or add to the DJ role list",
					Required:    true,
				},
			},
		},
		{
			Name:        "djroletoggle",
			Description: "Adds or removes the DJ role from the specified user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to add or remove the DJ role from",
					Required:    true,
				},
			},
		},
		{
			Name:        "webui",
			Description: "Generates a link to the web UI, which allows you to control the bot from a web browser.",
		},
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Plays a song from YouTube or spotify playlist. If it's not a valid link, it will insert into the queue the first result for the given queue
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = server[i.GuildID].PlayCommand(&clients, i, false, owner)
		},
		// Plays a playlist from YouTube or spotify (or searches the query on YouTube)
		"playlist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = server[i.GuildID].PlayCommand(&clients, i, true, owner)
		},
		// Skips the currently playing song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					el := server[i.GuildID].Queue.GetFirstElement()
					server[i.GuildID].Skip <- manager.Skip
					server[i.GuildID].Paused.Store(false)

					if server[i.GuildID].Queue.IsEmpty() {
						server[i.GuildID].ChanQuitVC <- true
					}

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SkipTitle,
						el.Title+" - "+el.Duration+" added by "+el.User).
						SetColor(0x7289DA).SetThumbnail(el.Thumbnail).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SkipTitle, constants.QueueEmpty).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},

		// Clears the entire queue
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					go server[i.GuildID].Clean()
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.QueueTitle, constants.QueueCleared).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.QueueTitle, constants.QueueEmpty).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			const maxQueue = 10

			if server[i.GuildID].IsPlaying() {
				el := server[i.GuildID].Queue.GetAllQueue()
				e := embed.NewEmbed().SetTitle(s.State.User.Username).SetDescription(constants.QueueTitle).AddField("1", fmt.Sprintf("[%s](%s) - %s/%s added by %s\n", el[0].Title, el[0].Link,
					manager.FormatDuration(float64(server[i.GuildID].Frames.Load())/constants.FrameSeconds), el[0].Duration, el[0].User))

				var nEl int
				if len(el) > maxQueue {
					nEl = maxQueue
				} else {
					nEl = len(el)
				}

				// Generate song info for the message
				for j := 1; j < nEl; j++ {
					e = e.AddField(strconv.Itoa(j+1), fmt.Sprintf("[%s](%s) - %s added by %s\n", el[j].Title, el[j].Link, el[j].Duration, el[j].User))
				}

				// Add the number of songs not shown if the queue is longer than maxQueue
				if len(el) > maxQueue {
					e = e.AddField("...", "And "+strconv.Itoa(len(el)-maxQueue)+" more")
				}

				// Send embed
				embed.SendAndDeleteEmbedInteraction(s, e.SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*20, nil)
			} else {
				// Queue is empty
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.QueueTitle, constants.QueueEmpty).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					if server[i.GuildID].Paused.CompareAndSwap(false, true) {
						server[i.GuildID].Pause <- struct{}{}
						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.PauseTitle, constants.Paused).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.PauseTitle, constants.AlreadyPaused).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.PauseTitle, constants.QueueEmpty).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					if server[i.GuildID].Paused.CompareAndSwap(true, false) {
						server[i.GuildID].Resume <- struct{}{}
						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ResumeTitle, constants.Resumed).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ResumeTitle, constants.AlreadyResumed).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ResumeTitle, constants.QueueEmpty).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			c := embed.DeferResponse(clients.Discord, i.Interaction)

			// Check if user is not in a voice channel
			if manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if !server[i.GuildID].IsPlaying() {
					server[i.GuildID].VC.Disconnect()
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DisconnectedTitle, constants.Disconnected).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.StillPlaying).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
			}
		},
		// Restarts the bot
		"restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the owner of the bot is the one who sent the command
			if owner == i.Member.User.ID {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.RestartTitle, constants.Disconnected).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*1, nil)

				_ = s.Close()
				clients.Database.Close()
				os.Exit(0)
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, "I'm sorry "+i.Member.User.Username+", I'm afraid I can't do that").SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		// Creates a custom command to play a song or playlist
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			var (
				command = strings.ToLower(options[0].Value.(string))
				song    = options[1].Value.(string)
				loop    = options[2].Value.(bool)
			)

			if server[i.GuildID].Custom[command] == nil {
				err := clients.Database.AddCommand(command, song, i.GuildID, loop)
				server[i.GuildID].Custom[command] = &database.CustomCommand{Link: song, Loop: loop}

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, err.Error()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SuccessfulTitle, constants.CommandAdded).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.CommandExists).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		// Removes a custom command from the DB
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if command := i.ApplicationCommandData().Options[0].Value.(string); server[i.GuildID].Custom[command] != nil {
				err := clients.Database.RemoveCustom(i.ApplicationCommandData().Options[0].Value.(string), i.GuildID)
				delete(server[i.GuildID].Custom, command)

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, err.Error()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SuccessfulTitle, constants.CommandRemoved).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.CommandNotExists).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		// Lists all custom commands for the current server
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commands := make([]string, 0, len(server[i.GuildID].Custom))

			for c := range server[i.GuildID].Custom {
				commands = append(commands, c)
			}

			sort.Strings(commands)

			embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.CommandsTitle, strings.Join(commands, ", ")).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30, nil)
		},
		// Calls a custom command
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			c := embed.DeferResponse(clients.Discord, i.Interaction)

			if server[i.GuildID].DjModeCheck(clients.Discord, i, owner, nil) {
				return
			}

			command := strings.ToLower(i.ApplicationCommandData().Options[0].Value.(string))

			if server[i.GuildID].Custom[command] != nil {
				// Check if user is not in a voice channel
				if vs := manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
					if manager.JoinVC(i.Interaction, vs.ChannelID, s, server[i.GuildID], c) {
						p := manager.PlayEvent{
							Username:    i.Member.User.Username,
							Song:        server[i.GuildID].Custom[command].Link,
							Clients:     &clients,
							Interaction: i.Interaction,
							Random:      false,
							Loop:        server[i.GuildID].Custom[command].Loop,
							Priority:    false,
							IsDeferred:  c,
						}

						if len(i.ApplicationCommandData().Options) > 1 {
							p.Priority = i.ApplicationCommandData().Options[1].Value.(bool)
						}
						server[i.GuildID].Play(p)
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.CommandInvalid).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
			}
		},
		// Stats, like latency, and the size of the local cache
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			size, files := manager.FolderStats(constants.CachePath)

			embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.StatsTitle, "Called by "+i.Member.User.Username).
				AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
				AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(files)+", "+
				manager.ByteCountSI(size)).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15, nil)
		},
		// Refreshes things about a song
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var (
				options = i.ApplicationCommandData().Options
				query   = options[0].Value.(string)
				info    = options[1].Value.(bool)
				song    = options[2].Value.(bool)
			)

			if manager.IsValidURL(query) {
				if el, err := clients.Database.CheckInDb(query); err != nil {
					// Check if it's a playlist
					if entries, err := clients.Database.GetPlaylist(query); err == nil && len(entries) > 0 {
						err := clients.Database.RemovePlaylist(query)
						if err != nil {
							lit.Error("Error while removing playlist from db: %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SuccessfulTitle,
							constants.UpdateQueued).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
					} else {
						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotCached).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
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

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SuccessfulTitle,
						constants.UpdateQueued).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				// Check if it's in the search results
				if search, err := clients.Database.GetSearch(query); err == nil && search != "" {
					err := clients.Database.RemoveSearch(query)
					if err != nil {
						lit.Error("Error while removing search from db: %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.SuccessfulTitle,
						constants.UpdateQueued).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.InvalidURL).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			}
		},
		"blacklist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member.User.ID == owner {
				if id := i.ApplicationCommandData().Options[0].UserValue(nil).ID; id == i.Member.User.ID {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
						"You are really trying to add yourself to the blacklist?").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
				} else {
					if _, ok := blacklist[id]; ok {
						// Removing from the blacklist
						delete(blacklist, id)

						err := clients.Database.RemoveFromBlacklist(id)
						if err != nil {
							lit.Error("Error while deleting from blacklist, %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.BlacklistTitle,
							constants.BlacklistRemoved).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
					} else {
						// Adding
						blacklist[id] = true

						err := clients.Database.AddToBlacklist(id)
						if err != nil {
							lit.Error("Error while inserting from blacklist, %s", err)
						}

						embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.BlacklistTitle,
							constants.BlacklistAdded).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
					}
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
			}
		},
		// Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...
		"goto": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if server[i.GuildID].IsPlaying() {
				t, err := time.ParseDuration(i.ApplicationCommandData().Options[0].Value.(string))
				if err != nil {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.GotoInvalid).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					server[i.GuildID].Queue.ModifyFirstElement(func(e *queue.Element) {
						if e.Segments == nil {
							e.Segments = make(map[int]bool)
						}

						server[i.GuildID].Paused.Store(true)
						server[i.GuildID].Pause <- struct{}{}

						e.Segments[int(server[i.GuildID].Frames.Load()+1)] = true
						e.Segments[int(t.Seconds()*constants.FrameSeconds)] = true

						server[i.GuildID].Resume <- struct{}{}
						server[i.GuildID].Paused.Store(false)
					})

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.GotoTitle, constants.SkippedTo+t.String()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NothingPlaying).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
		// Streams a song from the given URL, useful for radios
		"stream": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			c := embed.DeferResponse(clients.Discord, i.Interaction)
			if server[i.GuildID].DjModeCheck(s, i, owner, c) {
				return
			}

			if vs := manager.FindUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
				url := i.ApplicationCommandData().Options[0].Value.(string)
				if !strings.HasPrefix(url, "file") && manager.IsValidURL(url) {
					c := make(chan struct{})
					go embed.SendEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).
						AddField(constants.EnqueuedTitle, url).SetColor(0x7289DA).MessageEmbed, i.Interaction, c, c)

					stdout, cmds := manager.Stream(url)
					el := queue.Element{
						ID:          url,
						Title:       "Stream",
						Duration:    "NaN",
						Link:        url,
						User:        i.Member.User.Username,
						TextChannel: i.ChannelID,
						BeforePlay: func() {
							manager.CmdsStart(cmds)
						},
						AfterPlay: func() {
							manager.CmdsKill(cmds)
						},
						Reader: stdout,
						Closer: stdout,
					}

					if manager.JoinVC(i.Interaction, vs.ChannelID, s, server[i.GuildID], c) {
						go manager.DeleteInteraction(s, i.Interaction, c)
						if len(i.ApplicationCommandData().Options) > 1 {
							server[i.GuildID].AddSong(i.ApplicationCommandData().Options[1].Value.(bool), el)
						} else {
							server[i.GuildID].AddSong(false, el)
						}
					}
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.InvalidURL).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
			}
		},
		// Enables or disables DJ mode
		"dj": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member.User.ID == owner {
				if server[i.GuildID].DjMode {
					server[i.GuildID].DjMode = false
					err := clients.Database.SetDJSettings(i.GuildID, false)
					if err != nil {
						lit.Error("Error while disabling DJ mode, %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DjTitle, constants.DjDisabled).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					server[i.GuildID].DjMode = true
					err := clients.Database.SetDJSettings(i.GuildID, true)
					if err != nil {
						lit.Error("Error while enabling DJ mode, %s", err)
					}

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DjTitle, constants.DjEnabled).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
			}
		},
		// Sets the DJ role
		"djrole": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member.User.ID == owner {
				role := i.ApplicationCommandData().Options[0].RoleValue(nil, "")
				if role.ID != server[i.GuildID].DjRole {
					server[i.GuildID].DjRole = role.ID
					err := clients.Database.UpdateDJRole(i.GuildID, role.ID)
					if err != nil {
						lit.Error("Error updating DJ role: %s", err.Error())
					}

					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DjTitle, constants.DjRoleChanged).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DjTitle, constants.DjRoleEqual).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
			}
		},
		// Adds or removes the DJ role from a user
		"djroletoggle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member.User.ID == owner {
				var err error
				var action string

				user, _ := s.GuildMember(i.GuildID, i.ApplicationCommandData().Options[0].UserValue(nil).ID)
				if !manager.HasRole(user.Roles, server[i.GuildID].DjRole) {
					err = s.GuildMemberRoleAdd(i.GuildID, user.User.ID, server[i.GuildID].DjRole)
					action = "added!"
				} else {
					err = s.GuildMemberRoleRemove(i.GuildID, user.User.ID, server[i.GuildID].DjRole)
					action = "removed!"
				}

				if err != nil {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
						"The bot doesn't have the necessary permission!").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
				} else {
					embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.DjTitle,
						"The role has been succefully "+action).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
				}
			} else {
				embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle,
					"Only the owner of the bot can use this command!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, nil)
			}
		},
		// Generates a link to the web UI
		"webui": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if vs := manager.FindUserVoiceState(clients.Discord, i.GuildID, i.Member.User.ID); vs != nil {
				token := webApi.AddUser(i.Member.User, api.UserInfo{Guild: i.GuildID, TextChannel: i.ChannelID})
				embed := embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.WebUITitle, fmt.Sprintf("%s/?token=%s&GuildId=%s", origin, token, i.GuildID)).SetColor(0x7289DA).MessageEmbed

				// Send the response as ephemeral
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				})
			} else {
				embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, nil)
			}
		},
	}
)
