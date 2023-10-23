package manager

import (
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/bwmarrin/discordgo"
	"time"
)

func (server *Server) PlayCommand(clients *Clients, i *discordgo.InteractionCreate, playlist bool, owner string) (status PlayStatus) {
	if server.DjModeCheck(clients.Discord, i, owner) {
		return DjMode
	}

	// Check if user is not in a voice channel
	if vs := FindUserVoiceState(clients.Discord, i.GuildID, i.Member.User.ID); vs != nil {
		if JoinVC(i.Interaction, vs.ChannelID, clients.Discord, server) {
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
				server.Play(clients, link, i.Interaction, vs.GuildID, i.Member.User.Username, shuffle, loop, priority)
				status = Success
			} else {
				embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle,
					"Playlist detected, but playlist command not used.").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10)
				status = Playlist
			}
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.NotInVC).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
		status = NotInVC
	}

	return
}

func (server *Server) DjModeCheck(s *discordgo.Session, i *discordgo.InteractionCreate, owner string) bool {
	if server.DjMode && i.Member.User.ID != owner && !HasRole(i.Member.Roles, server.DjRole) {
		embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.DjNot).
			SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*3)
		return true
	}
	return false
}
