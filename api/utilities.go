package api

import "github.com/bwmarrin/discordgo"

func stringToBool(s string) bool {
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else {
		return false
	}
}

func (a *Api) checkAuthorization(token string) (*discordgo.User, bool) {
	if token == "" {
		return nil, false
	}

	var u *discordgo.User
	var ok bool
	if u, ok = a.tokensToUsers[token]; !ok {
		return nil, false
	}

	return u, true
}

func (a *Api) checkAuthorizationAndGuild(token, guild string) (*discordgo.User, bool) {
	u, ok := a.checkAuthorization(token)
	if !ok {
		return nil, false
	}

	if a.userInfo[a.tokensToUsers[token].ID].Guild != guild {
		return nil, false
	}
	return u, true
}

// Generates an interaction for the play command.
func (a *Api) interactionGenerator(u *discordgo.User, song string, playlist bool, shuffle bool, loop bool, priority bool, guild string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{
				User: u,
			},
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{
						Name:  "link",
						Value: song,
					},
					{
						Name:  "playlist",
						Value: playlist,
					},
					{
						Name:  "shuffle",
						Value: shuffle,
					},
					{
						Name:  "loop",
						Value: loop,
					},
					{
						Name:  "priority",
						Value: priority,
					},
				},
			},
			ChannelID: a.userInfo[u.ID].TextChannel,
			GuildID:   guild,
		},
	}
}
