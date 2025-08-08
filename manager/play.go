package manager

import (
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (server *Server) PlayCommand(clients *Clients, e *events.ApplicationCommandInteractionCreate, playlist bool, owner string) (status PlayStatus) {
	c := embed.DeferResponse(e)

	if server.DjModeCheck(e, owner, c) {
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
				link, err = filterPlaylist(link)
			}

			if err == nil {
				server.Play(PlayEvent{
					Username:   e.Member().User.Username,
					Song:       link,
					Clients:    clients,
					Event:      e,
					Random:     shuffle,
					Loop:       loop,
					Priority:   priority,
					IsDeferred: c,
				})

				status = Success
			} else {
				embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle,
					"Playlist detected, but playlist command not used.", false).
					SetColor(0x7289DA).Build(), e, time.Second*10, c)
				status = Playlist
			}
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.NotInVC, false).
			SetColor(0x7289DA).Build(), e, time.Second*5, c)
		status = NotInVC
	}

	return
}

func (server *Server) DjModeCheck(e *events.ApplicationCommandInteractionCreate, owner string, isDeferred chan struct{}) bool {
	if server.DjMode && e.Member().User.ID.String() != owner && !HasRole(e.Member().RoleIDs, server.DjRole) {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.DjNot, false).
			SetColor(0x7289DA).Build(), e, time.Second*3, isDeferred)
		return true
	}
	return false
}
