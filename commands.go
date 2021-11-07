package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io/ioutil"
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
			Name:        "lyrics",
			Description: "Gets lyrics for the a song, or if not specified for the currently playing one.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "song",
					Description: "Link to the playlist to play",
					Required:    false,
				},
			},
		},
		{
			Name:        "summon",
			Description: "Make the bot join your voice channel",
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
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Plays a song from YouTube or spotify playlist. If it's not a valid link, it will insert into the queue the first result for the given queue
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			play(s, i.ApplicationCommandData().Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, false)
		},

		// Skips the currently playing song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			server[i.GuildID].skip = true

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(skipTitle,
				server[i.GuildID].queue[0].title+" - "+server[i.GuildID].queue[0].duration+" added by "+server[i.GuildID].queue[0].user).
				SetColor(0x7289DA).SetThumbnail(server[i.GuildID].queue[0].thumbnail).MessageEmbed, i.Interaction, time.Second*5)
		},

		// Clears the entire queue
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			server[i.GuildID].clear = true
			server[i.GuildID].skip = true

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueCleared).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
		},

		// Inserts the song from the given playlist in a random order in the queue
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			play(s, i.ApplicationCommandData().Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, true)
		},

		// Pauses the currently playing song
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].queue) > 0 && !server[i.GuildID].isPaused {
				server[i.GuildID].isPaused = true
				go sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(pauseTitle, paused).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				server[i.GuildID].pause.Lock()
			}
		},

		// Resumes current song
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if server[i.GuildID].isPaused {
				server[i.GuildID].isPaused = false
				go sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(pauseTitle, resumed).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)

				server[i.GuildID].pause.Unlock()
				err := server[i.GuildID].vc.Speaking(true)
				if err != nil {
					lit.Error("vc.Speaking(true) failed: %s", err)
				}
			}
		},

		// Prints the currently playing song and the next songs
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].queue) > 0 {
				var message string

				// Generate song info for message
				for cont, el := range server[i.GuildID].queue {
					if cont == 0 {
						if el.title != "" {
							message += fmt.Sprintf("%d) [%s](%s) - %s/%s added by %s\n", cont+1, el.title, el.link,
								formatDuration(float64(server[i.GuildID].queue[0].frame/frameSeconds)), el.duration, el.user)
							continue
						} else {
							message += "Currently playing: Getting info...\n\n"
							continue
						}
					}

					// If we don't have the title, we use some placeholder text
					if el.title == "" {
						message += fmt.Sprintf("%d) Getting info...\n", cont+1)
					} else {
						message += fmt.Sprintf("%d) [%s](%s) - %s added by %s\n", cont+1, el.title, el.link, el.duration, el.user)
					}
				}

				// Send embed
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, message).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15)
			} else {
				// Queue is empty
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(queueTitle, queueEmpty).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Tries to search for lyrics of the specified song, or if not specified searches for the title of the currently playing song
		"lyrics": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// We search for lyrics only if there's something playing
			if len(server[i.GuildID].queue) > 0 {
				c := make(chan int)
				go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(lyricsTitle, searching).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)
				var song string

				// If the user didn't input a title, we use the currently playing video
				if len(i.ApplicationCommandData().Options) > 0 {
					song = i.ApplicationCommandData().Options[0].StringValue()
				} else {
					song = server[i.GuildID].queue[0].title
				}

				text := formatLongMessage(lyrics(song))

				<-c
				err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
				if err != nil {
					lit.Error("InteractionResponseDelete failed: %s", err.Error())
				}

				mex, err := s.ChannelMessageSend(i.ChannelID, "Lyrics for "+song+": ")
				if err != nil {
					lit.Error("Lyrics MessageSend failed: %s", err)
					return
				}

				server[i.GuildID].queue[0].messageID = append(server[i.GuildID].queue[0].messageID, *mex)

				// If the messages are more then 3, we don't send anything
				if len(text) > 3 {
					mex, _ := s.ChannelMessageSend(i.ChannelID, "```Lyrics too long!```")
					server[i.GuildID].queue[0].messageID = append(server[i.GuildID].queue[0].messageID, *mex)
					return
				}

				for _, t := range text {
					mex, _ = s.ChannelMessageSend(i.ChannelID, "```"+t+"```")
					server[i.GuildID].queue[0].messageID = append(server[i.GuildID].queue[0].messageID, *mex)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, errorNotPlaying).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Make the bot join your voice channel
		"summon": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			var err error

			// If we are playing something, we lock the pause mutex
			if len(server[i.GuildID].queue) > 0 {
				server[i.GuildID].pause.Lock()

				// Disconnect the bot
				if server[i.GuildID].vc != nil {
					_ = server[i.GuildID].vc.Disconnect()
				}

				// And reconnect the bot to the new voice channel
				server[i.GuildID].queue[0].channel = vs.ChannelID
				server[i.GuildID].vc, err = s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)

				server[i.GuildID].pause.Unlock()
			} else {
				// Else we just join the channel and wait
				server[i.GuildID].server.Lock()

				server[i.GuildID].vc, err = s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)

				// We also start the quitVC routine to disconnect the bot after a minute of inactivity
				go quitVC(i.GuildID)

				server[i.GuildID].server.Unlock()
			}

			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, cantJoinVC+err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				c, err := s.Channel(vs.ChannelID)
				if err == nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(summonTitle, joined+c.Name).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			}
		},

		// Disconnect the bot from the voice channel
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the queue is empty
			if len(server[i.GuildID].queue) == 0 {
				server[i.GuildID].server.Lock()

				_ = server[i.GuildID].vc.Disconnect()
				server[i.GuildID].vc = nil

				server[i.GuildID].server.Unlock()

				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(disconnectedTitle, disconnected).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, stillPlaying).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Restarts the bot
		"restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the owner of the bot required the restart
			if owner == i.Member.User.ID {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(restartTitle, disconnected).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*1)
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

		// Stats, like latency, and the size of the local cache
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			files, _ := ioutil.ReadDir("./audio_cache")

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(statsTitle, "Called by "+i.Member.User.Username).
				AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
				AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(len(files))+", "+
				ByteCountSI(DirSize("./audio_cache"))).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15)
		},

		// Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...
		"goto": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].queue) > 0 {
				t, err := time.ParseDuration(i.ApplicationCommandData().Options[0].StringValue())
				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, gotoInvalid).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					if server[i.GuildID].queue[0].segments == nil {
						server[i.GuildID].queue[0].segments = make(map[int]bool)
					}

					server[i.GuildID].pause.Lock()

					server[i.GuildID].queue[0].segments[server[i.GuildID].queue[0].frame+1] = true
					server[i.GuildID].queue[0].segments[int(t.Seconds()*frameSeconds)] = true

					server[i.GuildID].pause.Unlock()

					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(gotoTitle, skippedTo+t.String()).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, nothingPlaying).
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
				vs := findUserVoiceState(s, i.Interaction)

				// Check if user is not in a voice channel
				if vs == nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
					return
				}

				if server[i.GuildID].custom[command].loop {
					playLoop(s, i.Interaction, server[i.GuildID].custom[command].link)
				} else {
					play(s, server[i.GuildID].custom[command].link, i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, false)
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, commandInvalid).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Streams a song from the given URL, useful for radios
		"stream": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var (
				url = i.ApplicationCommandData().Options[0].StringValue()
				vs  = findUserVoiceState(s, i.Interaction)
			)

			if strings.HasPrefix(url, "file") || !isValidURL(url) {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, invalidURL).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notInVC).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			c := make(chan int)
			go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(enqueuedTitle, url).SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

			stdout, cmds := stream(url)
			el := Queue{url, "NaN", url, url, i.Member.User.Username, nil, "", 0, nil, i.ChannelID}

			// Adds to queue
			server[i.GuildID].queueMutex.Lock()
			server[i.GuildID].queue = append(server[i.GuildID].queue, el)
			server[i.GuildID].queueMutex.Unlock()

			// Starts command and plays URL
			playSound(s, i.GuildID, vs.ChannelID, url, i.Interaction, stdout, &c, cmds)

			// After we have finished, closes pipe and unlocks mutex
			_ = stdout.Close()
			server[i.GuildID].stream.Unlock()
		},

		// Loops a song from the url
		"loop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			playLoop(s, i.Interaction, i.ApplicationCommandData().Options[0].StringValue())
		},

		// Refreshes things about a song
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			url := i.ApplicationCommandData().Options[0].StringValue()
			info := i.ApplicationCommandData().Options[1].BoolValue()
			song := i.ApplicationCommandData().Options[2].BoolValue()

			if !isValidURL(url) {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, invalidURL).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			el := checkInDb(url)
			if el.title == "" {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, notCached).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			if info {
				removeFromDB(el)
			}

			if song {
				err := os.Remove(cachePath + el.id + audioExtension)
				if err != nil {
					lit.Error(err.Error())
				}
			}

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(successfulTitle,
				"Requested data will be updated next time the song is played!").
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
		},
	}
)
