package api

import (
	"github.com/disgoorg/disgo/discord"
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

func (a *Api) checkAuthorization(token string) (*discord.Member, bool) {
	if token == "" {
		return nil, false
	}

	var u *discord.Member
	var ok bool
	if u, ok = a.tokensToUsers[token]; !ok {
		return nil, false
	}

	return u, true
}

func (a *Api) checkAuthorizationAndGuild(token, guild string) (*discord.Member, bool) {
	u, ok := a.checkAuthorization(token)
	if !ok {
		return nil, false
	}

	if a.userInfo[a.tokensToUsers[token].User.ID.String()].Guild != guild {
		return nil, false
	}
	return u, true
}
