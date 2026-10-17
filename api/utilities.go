package api

import (
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func stringToBool(s string) bool {
	return s == "true"
}

func (a *Api) checkAuthorization(token string) (*discord.Member, bool) {
	if token == "" {
		return nil, false
	}

	value, ok := a.tokensToUsers.Load(token)
	if !ok {
		return nil, false
	}

	return value.(*discord.Member), true
}

func (a *Api) checkAuthorizationAndGuild(token string, guild snowflake.ID) (*discord.Member, bool) {
	u, ok := a.checkAuthorization(token)
	if !ok {
		return nil, false
	}

	value, ok := a.userInfo.Load(u.User.ID)
	if !ok {
		return nil, false
	}

	if value.(*UserInfo).Guild != guild {
		return nil, false
	}

	return u, true
}

func (a *Api) getUserInfo(userID snowflake.ID) (*UserInfo, bool) {
	value, ok := a.userInfo.Load(userID)
	if !ok {
		return nil, false
	}

	return value.(*UserInfo), true
}

// checkOrigin verifies that the request origin matches the allowed origin
func (a *Api) checkOrigin(r *http.Request) bool {
	origin := a.origin

	if origin == "" || origin == "*" {
		return true
	}

	return r.Header.Get("Origin") == origin
}
