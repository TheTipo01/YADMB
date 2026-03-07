package manager

import (
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func (server *Server) PlayCommand(clients *Clients, e *events.ApplicationCommandInteractionCreate, playlist bool, owner map[snowflake.ID]struct{}) (status PlayStatus) {
	c := embed.DeferResponse(e)

	if server.DjModeCheck(e.Member().Member, owner) {
		return DjMode
	}

	// Check if user is not in a voice channel
	if vs := FindUserVoiceState(e.Client(), e.Member().GuildID, e.Member().User.ID); vs != nil {
		if JoinVC(e, *vs.ChannelID, server, c) {
			var (
				shuffle, loop, priority bool
				link                    string
				options                 = e.SlashCommandInteractionData()
			)

			shuffle = options.Bool("shuffle")
			loop = options.Bool("loop")
			priority = options.Bool("priority")

			var err error
			link = options.String("link")
			if !playlist {
				link, err = FilterPlaylist(link)
			}

			if err == nil {
				server.Play(PlayEvent{
					Username:    e.Member().User.Username,
					Song:        link,
					Clients:     clients,
					Event:       e,
					Random:      shuffle,
					Loop:        loop,
					Priority:    priority,
					IsDeferred:  c,
					TextChannel: e.Channel().ID(),
				})

				status = Success
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(constants.ErrorTitle,
					"Playlist detected, but playlist command not used.", false).
					WithColor(0x7289DA), e, time.Second*10, c)
				status = Playlist
			}
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
			WithColor(0x7289DA), e, time.Second*5, c)
		status = NotInVC
	}

	return
}

func (server *Server) DjModeCheck(member discord.Member, owner map[snowflake.ID]struct{}) bool {
	if _, isOwner := owner[member.User.ID]; server.DjMode && isOwner && !HasRole(member.RoleIDs, server.DjRole) {
		return true
	}
	return false
}
