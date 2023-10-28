package manager

import (
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/bwmarrin/discordgo"
	"time"
)

func (server *Server) PlayCommand(clients *Clients, i *discordgo.InteractionCreate, playlist bool, owner string) (status PlayStatus) {
	c := embed.DeferResponse(clients.Discord, i.Interaction)

	if server.DjModeCheck(clients.Discord, i, owner, c) {
		return DjMode
	}

	// Check if user is not in a voice channel
	if vs := FindUserVoiceState(clients.Discord, i.GuildID, i.Member.User.ID); vs != nil {
		if JoinVC(i.Interaction, vs.ChannelID, clients.Discord, server, c) {
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
				server.Play(PlayEvent{
					Username:    i.Member.User.Username,
					Song:        link,
					Clients:     clients,
					Interaction: i.Interaction,
					Random:      shuffle,
					Loop:        loop,
					Priority:    priority,
					IsDeferred:  c,
				})

				status = Success
			} else {
				embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle,
					"Playlist detected, but playlist command not used.").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10, c)
				status = Playlist
			}
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5, c)
		status = NotInVC
	}

	return
}

func (server *Server) DjModeCheck(s *discordgo.Session, i *discordgo.InteractionCreate, owner string, isDeferred chan struct{}) bool {
	if server.DjMode && i.Member.User.ID != owner && !HasRole(i.Member.Roles, server.DjRole) {
		embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.DjNot).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3, isDeferred)
		return true
	}
	return false
}
