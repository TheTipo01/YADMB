package main

import (
	"fmt"
	"github.com/TheTipo01/YADMB/database"
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
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Plays a song from YouTube or spotify playlist. If it's not a valid link, it will insert into the queue the first result for the given queue
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			playCommand(s, i, false)
		},
		// Plays a playlist from YouTube or spotify (or searches the query on YouTube)
		"playlist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			playCommand(s, i, true)
		},
		// Skips the currently playing song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil && server[i.GuildID].IsPlaying() {
				el := server[i.GuildID].queue.GetFirstElement()
				server[i.GuildID].skip <- struct{}{}
				server[i.GuildID].paused.Store(false)
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(skipTitle,
					el.Title+" - "+el.Duration+" added by "+el.User).
					SetColor(0x7289DA).SetThumbnail(el.Thumbnail).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Clears the entire queue
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].IsPlaying() {
					go server[i.GuildID].Clear()
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueCleared).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueEmpty).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			const maxQueue = 10

			if server[i.GuildID].IsPlaying() {
				el := server[i.GuildID].queue.GetAllQueue()
				embed := NewEmbed().SetTitle(s.State.User.Username).SetDescription(queueTitle).AddField("1", fmt.Sprintf("[%s](%s) - %s/%s added by %s\n", el[0].Title, el[0].Link,
					formatDuration(float64(server[i.GuildID].frames)/frameSeconds), el[0].Duration, el[0].User))

				var max int
				if len(el) > maxQueue {
					max = maxQueue
				} else {
					max = len(el)
				}

				// Generate song info for the message
				for j := 1; j < max; j++ {
					embed = embed.AddField(strconv.Itoa(j+1), fmt.Sprintf("[%s](%s) - %s added by %s\n", el[j].Title, el[j].Link, el[j].Duration, el[j].User))
				}

				// Add the number of songs not shown if the queue is longer than maxQueue
				if len(el) > maxQueue {
					embed = embed.AddField("...", "And "+strconv.Itoa(len(el)-maxQueue)+" more")
				}

				// Send embed
				sendAndDeleteEmbedInteraction(s, embed.SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*20)
			} else {
				// Queue is empty
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueEmpty).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].paused.CompareAndSwap(false, true) {
					server[i.GuildID].pause <- struct{}{}
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(pauseTitle, paused).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(pauseTitle, alreadyPaused).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if server[i.GuildID].paused.CompareAndSwap(true, false) {
					server[i.GuildID].resume <- struct{}{}
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(resumeTitle, resumed).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(resumeTitle, alreadyResumed).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.GuildID, i.Member.User.ID) != nil {
				if !server[i.GuildID].IsPlaying() {
					_ = server[i.GuildID].vc.Disconnect()
					server[i.GuildID].vc = nil
					server[i.GuildID].voiceChannel = ""
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(disconnectedTitle, disconnected).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, stillPlaying).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Restarts the bot
		"restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the owner of the bot is the one who sent the command
			if owner == i.Member.User.ID {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(restartTitle, disconnected).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*1)

				_ = s.Close()
				db.Close()
				os.Exit(0)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, "I'm sorry "+i.Member.User.Username+", I'm afraid I can't do that").SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
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

			if server[i.GuildID].custom[command] == nil {
				err := db.AddCommand(command, song, i.GuildID, loop)
				server[i.GuildID].custom[command] = &database.CustomCommand{Link: song, Loop: loop}

				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle, commandAdded).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, commandExists).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Removes a custom command from the DB
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if command := i.ApplicationCommandData().Options[0].Value.(string); server[i.GuildID].custom[command] != nil {
				err := db.RemoveCustom(i.ApplicationCommandData().Options[0].Value.(string), i.GuildID)
				delete(server[i.GuildID].custom, command)

				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle, commandRemoved).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, commandNotExists).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Lists all custom commands for the current server
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commands := make([]string, 0, len(server[i.GuildID].custom))

			for c := range server[i.GuildID].custom {
				commands = append(commands, c)
			}

			sort.Strings(commands)

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(commandsTitle, strings.Join(commands, ", ")).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30)
		},
		// Calls a custom command
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := strings.ToLower(i.ApplicationCommandData().Options[0].Value.(string))

			if server[i.GuildID].custom[command] != nil {
				// Check if user is not in a voice channel
				if vs := findUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
					if joinVC(i.Interaction, vs.ChannelID) {
						if len(i.ApplicationCommandData().Options) > 1 {
							play(s, server[i.GuildID].custom[command].Link, i.Interaction, vs.GuildID, i.Member.User.Username, false, server[i.GuildID].custom[command].Loop, i.ApplicationCommandData().Options[1].Value.(bool))
						} else {
							play(s, server[i.GuildID].custom[command].Link, i.Interaction, vs.GuildID, i.Member.User.Username, false, server[i.GuildID].custom[command].Loop, false)
						}
					}
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, commandInvalid).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Stats, like latency, and the size of the local cache
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			size, files := FolderStats(cachePath)

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(statsTitle, "Called by "+i.Member.User.Username).
				AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
				AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(files)+", "+
				ByteCountSI(size)).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15)
		},
		// Refreshes things about a song
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var (
				options = i.ApplicationCommandData().Options
				url     = options[0].Value.(string)
				info    = options[1].Value.(bool)
				song    = options[2].Value.(bool)
			)

			if isValidURL(url) {
				if el, err := db.CheckInDb(url); err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notCached).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					if info {
						db.RemoveFromDB(el)
					}

					if song {
						err := os.Remove(cachePath + el.ID + audioExtension)
						if err != nil {
							lit.Error(err.Error())
						}
					}

					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle,
						"Requested data will be updated next time the song is played!").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, invalidURL).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"blacklist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member.User.ID == owner {
				if id := i.ApplicationCommandData().Options[0].UserValue(nil).ID; id == i.Member.User.ID {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle,
						"You are really trying to add yourself to the blacklist?").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
				} else {
					if _, ok := blacklist[id]; ok {
						// Removing from the blacklist
						delete(blacklist, id)

						err := db.RemoveFromBlacklist(id)
						if err != nil {
							lit.Error("Error while deleting from blacklist, %s", err)
						}

						sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(blacklistTitle,
							"User removed from the blacklist!").
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
					} else {
						// Adding
						blacklist[id] = true

						err := db.AddToBlacklist(id)
						if err != nil {
							lit.Error("Error while inserting from blacklist, %s", err)
						}

						sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(blacklistTitle,
							"User added to the blacklist!").
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
					}
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle,
					"Only the owner of the bot can use this command!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
			}
		},
		// Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...
		"goto": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if server[i.GuildID].IsPlaying() {
				t, err := time.ParseDuration(i.ApplicationCommandData().Options[0].Value.(string))
				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, gotoInvalid).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					server[i.GuildID].queue.ModifyFirstElement(func(e *queue.Element) {
						if e.Segments == nil {
							e.Segments = make(map[int]bool)
						}

						server[i.GuildID].paused.Store(true)
						server[i.GuildID].pause <- struct{}{}

						e.Segments[server[i.GuildID].frames+1] = true
						e.Segments[int(t.Seconds()*frameSeconds)] = true

						server[i.GuildID].resume <- struct{}{}
						server[i.GuildID].paused.Store(false)
					})

					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(gotoTitle, skippedTo+t.String()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, nothingPlaying).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Streams a song from the given URL, useful for radios
		"stream": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if vs := findUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
				url := i.ApplicationCommandData().Options[0].Value.(string)
				if !strings.HasPrefix(url, "file") && isValidURL(url) {
					c := make(chan struct{})
					go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(enqueuedTitle, url).SetColor(0x7289DA).MessageEmbed, i.Interaction, c)

					stdout, cmds := stream(url)
					el := queue.Element{
						ID:          url,
						Title:       "Stream",
						Duration:    "NaN",
						Link:        url,
						User:        i.Member.User.Username,
						TextChannel: i.ChannelID,
						BeforePlay: func() {
							cmdsStart(cmds)
						},
						AfterPlay: func() {
							cmdsKill(cmds)
						},
						Reader: stdout,
						Closer: stdout,
					}

					if joinVC(i.Interaction, vs.ChannelID) {
						go deleteInteraction(s, i.Interaction, c)
						if len(i.ApplicationCommandData().Options) > 1 {
							server[i.GuildID].AddSong(i.ApplicationCommandData().Options[1].Value.(bool), el)
						} else {
							server[i.GuildID].AddSong(false, el)
						}
					}
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, invalidURL).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
	}
)

func playCommand(s *discordgo.Session, i *discordgo.InteractionCreate, playlist bool) {
	// Check if user is not in a voice channel
	if vs := findUserVoiceState(s, i.GuildID, i.Member.User.ID); vs != nil {
		if joinVC(i.Interaction, vs.ChannelID) {
			var (
				shuffle, loop, priority bool
				link                    string
				options                 = i.ApplicationCommandData().Options
			)

			for j := 1; j < len(options); j++ {
				switch options[j].Name {
				case "shuffle":
					shuffle = options[j].Value.(bool)
				case "loop":
					loop = options[j].Value.(bool)
				case "priority":
					priority = options[j].Value.(bool)
				}
			}

			var err error
			if playlist {
				link = options[0].Value.(string)
			} else {
				link, err = filterPlaylist(options[0].Value.(string))
			}

			if err == nil {
				play(s, link, i.Interaction, vs.GuildID, i.Member.User.Username, shuffle, loop, priority)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle,
					"Playlist detected, but playlist command not used.").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10)
			}
		}
	} else {
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
	}
}
