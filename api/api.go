package api

import (
	"github.com/TheTipo01/YADMB/manager"
	"github.com/bwmarrin/discordgo"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

func NewApi(servers map[string]*manager.Server, address, owner string, clients *manager.Clients) Api {
	r := gin.New()
	r.Use(gin.Recovery())

	a := Api{servers: servers, tokensToUsers: make(map[string]*discordgo.User), userInfo: make(map[string]*UserInfo), owner: owner, clients: clients}

	r.GET("/queue/:guild", a.getQueue)
	r.POST("/queue/:guild", a.addToQueue)
	r.DELETE("/queue/:guild", a.skip)
	r.GET("/song/resume/:guild", a.resume)
	r.GET("/song/pause/:guild", a.pause)
	r.GET("/song/toggle/:guild", a.toggle)
	r.GET("/favorites", a.getFavorites)
	r.POST("/favorites", a.addFavorite)
	r.DELETE("/favorites", a.removeFavorite)
	go r.Run(address)

	return a
}

// AddUser adds a user to the api, returning the token.
// If the user is already in the api, it removes it and generates a new one.
func (a *Api) AddUser(user *discordgo.User, userInfo UserInfo) string {
	if u, ok := a.userInfo[user.ID]; ok {
		delete(a.tokensToUsers, u.token)

		if a.userInfo[user.ID].LongLivedToken == "" {
			delete(a.userInfo, user.ID)
		}
	}

	// Generate a new token until it is unique
	var token string
	for {
		token = uniuri.NewLen(32)
		if _, ok := a.tokensToUsers[token]; !ok {
			break
		}
	}

	a.tokensToUsers[token] = user

	if a.userInfo[user.ID] != nil {
		userInfo.LongLivedToken = a.userInfo[user.ID].LongLivedToken
	}
	userInfo.token = token

	a.userInfo[user.ID] = &userInfo

	return token
}

func (a *Api) AddLongLivedToken(user *discordgo.User, userInfo UserInfo) {
	a.tokensToUsers[userInfo.LongLivedToken] = user

	if a.userInfo[user.ID] != nil {
		userInfo.token = a.userInfo[user.ID].token
	}
	a.userInfo[user.ID] = &userInfo
}
