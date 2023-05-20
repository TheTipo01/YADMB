package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
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
			Name:        "preload",
			Description: "Preloads a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link of the song to preload",
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
			if vs != nil {
				if server[i.GuildID].vc == nil {
					// Join the voice channel
					vc, err := s.ChannelVoiceJoin(i.GuildID, vs.ChannelID, false, true)
					if err != nil {
						lit.Error("Error joining voice channel: %s", err.Error())
						return
					}
					server[i.GuildID].vc = vc
				}

				// If the user requested a playlist, don't remove the parameter
				if len(i.ApplicationCommandData().Options) > 1 && i.ApplicationCommandData().Options[1].BoolValue() {
					play(s, i.ApplicationCommandData().Options[0].StringValue(), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.ID, false)
				} else {
					play(s, removePlaylist(i.ApplicationCommandData().Options[0].StringValue()), i.Interaction, vs.ChannelID, vs.GuildID, i.Member.User.ID, false)
				}
			}
		},

		// Skips the currently playing song
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) != nil && server[i.GuildID].queue.GetFirstElement() != nil {
				server[i.GuildID].skip = true
			}
		},

		// Clears the entire queue
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Check if user is not in a voice channel
			if findUserVoiceState(s, i.Interaction) != nil {
				server[i.GuildID].Clear()
			}
		},
	}
)
