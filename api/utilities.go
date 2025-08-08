package api

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func stringToBool(s string) bool {
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else {
		return false
	}
}

func (a *Api) checkAuthorization(token string) (*discord.User, bool) {
	if token == "" {
		return nil, false
	}

	var u *discord.User
	var ok bool
	if u, ok = a.tokensToUsers[token]; !ok {
		return nil, false
	}

	return u, true
}

func (a *Api) checkAuthorizationAndGuild(token, guild string) (*discord.User, bool) {
	u, ok := a.checkAuthorization(token)
	if !ok {
		return nil, false
	}

	if a.userInfo[a.tokensToUsers[token].ID.String()].Guild != guild {
		return nil, false
	}
	return u, true
}

// Generates an interaction for the play command.
func (a *Api) interactionGenerator(u *discord.User, song string, playlist bool, shuffle bool, loop bool, priority bool, guild string) *events.ApplicationCommandInteractionCreate {
	// TODO: Implement interaction generation logic
	return &events.ApplicationCommandInteractionCreate{}
}
