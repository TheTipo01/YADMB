package main

import (
	"fmt"
	"github.com/TheTipo01/YADMB/Queue"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// Commands
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Plays a song from youtube or spotify playlist (or search the query on youtube)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link or query to play",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "playlist",
					Description: "If the link is a playlist",
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
			Name:        "shuffle",
			Description: "Shuffles the songs in the playlist and adds them to the queue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link to the playlist to play",
					Required:    true,
				},
			},
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
			},
		},
		{
			Name:        "loop",
			Description: "Loops a song until skipped",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link to play in loop",
					Required:    true,
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
			// Check if user is not in a voice channel
			if vs := findUserVoiceState(s, i.Interaction); vs != nil {
				if joinVC(i.Interaction, vs.ChannelID) {
					// If the user requested a playlist, don't remove the parameter
					if len(i.ApplicationCommandData().Options) > 1 && i.ApplicationCommandData().Options[1].BoolValue() {
						play(s, i.ApplicationCommandData().Options[0].StringValue(), i.Interaction, vs.GuildID, i.Member.User.Username, false, false)
					} else {
						play(s, removePlaylist(i.ApplicationCommandData().Options[0].StringValue()), i.Interaction, vs.GuildID, i.Member.User.Username, false, false)
					}
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Skips the currently playing song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) != nil && !server[i.GuildID].queue.IsEmpty() {
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
			if findUserVoiceState(s, i.Interaction) != nil {
				if !server[i.GuildID].queue.IsEmpty() {
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
			if !server[i.GuildID].queue.IsEmpty() {
				var message string
				el := server[i.GuildID].queue.GetAllQueue()

				if el[0].Title != "" {
					message += fmt.Sprintf("%d) [%s](%s) - %s/%s added by %s\n", 1, el[0].Title, el[0].Link,
						formatDuration(float64(server[i.GuildID].frames)/frameSeconds), el[0].Duration, el[0].User)
				} else {
					message += "Currently playing: Getting info...\n\n"
				}

				// Generate song info for message
				for j := 1; j < len(el); j++ {
					// If we don't have the title, we use some placeholder text
					if el[j].Title == "" {
						message += fmt.Sprintf("%d) Getting info...\n", j+1)
					} else {
						message += fmt.Sprintf("%d) [%s](%s) - %s added by %s\n", j+1, el[j].Title, el[j].Link, el[j].Duration, el[j].User)
					}
				}

				// Send embed
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, message).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*20)
			} else {
				// Queue is empty
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueEmpty).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if findUserVoiceState(s, i.Interaction) != nil {
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
			if findUserVoiceState(s, i.Interaction) != nil {
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
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if vs := findUserVoiceState(s, i.Interaction); vs != nil {
				if joinVC(i.Interaction, vs.ChannelID) {
					play(s, i.ApplicationCommandData().Options[0].StringValue(), i.Interaction, vs.GuildID, i.Member.User.Username, true, false)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) != nil {
				if server[i.GuildID].queue.IsEmpty() {
					_ = server[i.GuildID].vc.Disconnect()
					server[i.GuildID].vc = nil
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
				_ = db.Close()
				os.Exit(0)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, "I'm sorry "+i.Member.User.Username+", I'm afraid I can't do that").SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Creates a custom command to play a song or playlist
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := addCommand(strings.ToLower(i.ApplicationCommandData().Options[0].StringValue()), i.ApplicationCommandData().Options[1].StringValue(), i.GuildID, i.ApplicationCommandData().Options[2].BoolValue())
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle, commandAdded).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Removes a custom command from the DB
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := removeCustom(i.ApplicationCommandData().Options[0].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle, commandRemoved).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
		// Lists all custom commands for the current server
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""

			for c := range server[i.GuildID].custom {
				message += c + ", "
			}

			message = message[:len(message)-2]

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(commandsTitle, message).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30)
		},
		// Calls a custom command
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())

			if server[i.GuildID].custom[command] != nil {
				// Check if user is not in a voice channel
				if vs := findUserVoiceState(s, i.Interaction); vs != nil {
					if joinVC(i.Interaction, vs.ChannelID) {
						play(s, server[i.GuildID].custom[command].link, i.Interaction, vs.GuildID, i.Member.User.Username, false, server[i.GuildID].custom[command].loop)
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
			files, _ := os.ReadDir("./audio_cache")

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(statsTitle, "Called by "+i.Member.User.Username).
				AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
				AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(len(files))+", "+
				ByteCountSI(DirSize("./audio_cache"))).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15)
		},
		// Refreshes things about a song
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			url := i.ApplicationCommandData().Options[0].StringValue()
			info := i.ApplicationCommandData().Options[1].BoolValue()
			song := i.ApplicationCommandData().Options[2].BoolValue()

			if isValidURL(url) {
				if el, err := checkInDb(url); err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notCached).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					if info {
						removeFromDB(el)
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

						_, err := db.Exec("DELETE FROM blacklist WHERE id=?", id)
						if err != nil {
							lit.Error("Error while deleting from blacklist, %s", err)
						}

						sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(blacklistTitle,
							"User removed from the blacklist!").
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
					} else {
						// Adding
						blacklist[id] = true

						_, err := db.Exec("INSERT INTO blacklist (`id`) VALUES(?)", id)
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
			if !server[i.GuildID].queue.IsEmpty() {
				t, err := time.ParseDuration(i.ApplicationCommandData().Options[0].StringValue())
				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, gotoInvalid).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					server[i.GuildID].queue.ModifyFirstElement(func(e *Queue.Element) {
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
			if vs := findUserVoiceState(s, i.Interaction); vs != nil {
				url := i.ApplicationCommandData().Options[0].StringValue()
				if !strings.HasPrefix(url, "file") && isValidURL(url) {
					c := make(chan int)
					go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(enqueuedTitle, url).SetColor(0x7289DA).MessageEmbed, i.Interaction, c)

					stdout, cmds := stream(url)
					el := Queue.Element{
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
						server[i.GuildID].AddSong(el)
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
		"loop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if vs := findUserVoiceState(s, i.Interaction); vs != nil {
				if joinVC(i.Interaction, vs.ChannelID) {
					play(s, removePlaylist(i.ApplicationCommandData().Options[0].StringValue()), i.Interaction, vs.GuildID, i.Member.User.Username, false, true)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
	}
)
