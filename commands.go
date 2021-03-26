package main

import (
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
					Name:        "customCommand",
					Description: "Name of the custom command",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link to the song/playlist",
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
					Name:        "customCommand",
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
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Plays a song
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			play(s, i.Data.Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, false)
		},

		// Skips a song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			server[i.GuildID].skip = true

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Skipped",
				server[i.GuildID].queue[0].title+" - "+server[i.GuildID].queue[0].duration+" added by "+server[i.GuildID].queue[0].user).
				SetColor(0x7289DA).SetThumbnail(server[i.GuildID].queue[0].thumbnail).MessageEmbed, i.Interaction, time.Second*5)
		},

		// Clear the queue of the guild
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			server[i.GuildID].clear = true
			server[i.GuildID].skip = true

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", "Queue cleared!").
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
		},

		// Randomly plays a song (or a playlist)
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				return
			}

			play(s, i.Data.Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, true)
		},

		// Pause the song
		"pause": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].queue) > 0 && !server[i.GuildID].isPaused {
				server[i.GuildID].isPaused = true
				go sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Paused the current song").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				server[i.GuildID].pause.Lock()
			}
		},

		// Resume playing
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if server[i.GuildID].isPaused {
				server[i.GuildID].isPaused = false
				go sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Pause", "Resumed the current song").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)

				server[i.GuildID].pause.Unlock()
				err := server[i.GuildID].vc.Speaking(true)
				if err != nil {
					lit.Error("vc.Speaking(true) failed: %s", err)
				}
			}
		},

		// Prints out queue for the guild
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var message string

			if len(server[i.GuildID].queue) > 0 {
				// Generate song info for message
				for cont, el := range server[i.GuildID].queue {
					if cont == 0 {
						if el.title != "" {
							message += "Currently playing: " + el.title + " - " + formatDuration(float64(server[i.GuildID].queue[0].frame/frameSeconds)) +
								"/" + el.duration + " added by " + el.user + "\n\n"
							continue
						} else {
							message += "Currently playing: Getting info...\n\n"
							continue
						}

					}
					// If we don't have the title, we use some placeholder text
					if el.title == "" {
						message += strconv.Itoa(cont) + ") Getting info...\n"
					} else {
						message += strconv.Itoa(cont) + ") " + el.title + " - " + el.duration + " by " + el.user + "\n"
					}
				}

				// Send embed
				em, err := s.ChannelMessageSendEmbed(i.ChannelID, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", message).
					SetColor(0x7289DA).MessageEmbed)
				if err != nil {
					lit.Error("Error sending queue embed: %s", err)
					return
				}

				// Wait for 15 seconds, then delete the message
				time.Sleep(time.Second * 15)
				err = s.ChannelMessageDelete(i.ChannelID, em.ID)
				if err != nil {
					lit.Error("Error deleting queue embed: %s", err)
				}
			} else {
				// Queue is empty
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Queue", "Queue is empty!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Prints lyrics of a song
		"lyrics": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// We search for lyrics only if there's something playing
			if len(server[i.GuildID].queue) > 0 {
				sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Lyrics", "Searching...").
					SetColor(0x7289DA).MessageEmbed, i.Interaction)
				var song string

				// If the user didn't input a title, we use the currently playing video
				if len(i.Data.Options) > 0 {
					song = i.Data.Options[0].StringValue()
				} else {
					song = server[i.GuildID].queue[0].title
				}

				text := formatLongMessage(lyrics(song))

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
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "To use this command something needs to be playing!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// We summon the bot in the user current voice channel
		"summon": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			vs := findUserVoiceState(s, i.Interaction)

			// Check if user is not in a voice channel
			if vs == nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
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
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't join voice channel!\n"+err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				c, err := s.Channel(vs.ChannelID)
				if err == nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Summon", "Joined "+c.Name).
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				}
			}
		},

		// Disconnect the bot from the guild voice channel
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the queue is empty
			if len(server[i.GuildID].queue) == 0 {
				server[i.GuildID].server.Lock()

				_ = server[i.GuildID].vc.Disconnect()
				server[i.GuildID].vc = nil

				server[i.GuildID].server.Unlock()

				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Disconnected", "Bye bye!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't disconnect the bot!\nStill playing in a voice channel.").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Makes the bot exit
		"restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if the owner of the bot required the restart
			if owner == i.Member.User.ID {
				os.Exit(0)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "I'm sorry "+i.Member.User.Username+", I'm afraid I can't do that").SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Adds a custom command
		"addcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := addCommand(strings.ToLower(i.Data.Options[0].StringValue()), i.Data.Options[1].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Successful", "Custom command added!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Removes a custom command
		"rmcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := removeCustom(i.Data.Options[0].StringValue(), i.GuildID)
			if err != nil {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", err.Error()).
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Successful", "Command removed successfully!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// Stats™
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			files, _ := ioutil.ReadDir("./audio_cache")

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Stats™", "Called by "+i.Member.User.Username).
				AddField("Latency", s.HeartbeatLatency().String()).AddField("Guilds", strconv.Itoa(len(s.State.Guilds))).
				AddField("Shard", strconv.Itoa(s.ShardID+1)+"/"+strconv.Itoa(s.ShardCount)).AddField("Cached song", strconv.Itoa(len(files))+", "+
				ByteCountSI(DirSize("./audio_cache"))).SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*15)
		},

		// Skips to a given time
		"goto": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(server[i.GuildID].queue) > 0 {
				t, err := time.ParseDuration(i.Data.Options[0].StringValue())
				if err != nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Wrong format.\nValid formats are: 1h10m3s, 3m, 4m10s...").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
				} else {
					if server[i.GuildID].queue[0].segments == nil {
						server[i.GuildID].queue[0].segments = make(map[int]bool)
					}

					server[i.GuildID].pause.Lock()

					server[i.GuildID].queue[0].segments[server[i.GuildID].queue[0].frame+1] = true
					server[i.GuildID].queue[0].segments[int(t.Seconds()*frameSeconds)] = true

					server[i.GuildID].pause.Unlock()
				}
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No songs playing!").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},

		// List custom commands
		"listcustom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""

			for c := range server[i.GuildID].custom {
				message += c + ", "
			}

			message = message[:len(message)-2]

			sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Commands", message).
				SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*30)
		},

		// Plays a custom commands
		"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			command := strings.ToLower(i.Data.Options[0].StringValue())

			if server[i.GuildID].custom[command] != "" {

				vs := findUserVoiceState(s, i.Interaction)

				// Check if user is not in a voice channel
				if vs == nil {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "You're not in a voice channel in this guild!").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
					return
				}

				play(s, i.Data.Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.Username, false)
			} else {
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Not a valid custom command!\nSee /listcustom for a list of custom commands.").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
			}
		},
	}
)
