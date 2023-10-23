package api

import (
	"github.com/TheTipo01/YADMB/manager"
	"github.com/bwmarrin/discordgo"
)

type Api struct {
	// Server managers
	servers map[string]*manager.Server
	// Map from token to users
	tokensToUsers map[string]*discordgo.User
	// Map from userID to token
	userInfo map[string]*UserInfo
	// Bot owner
	owner string
	// CLients for interacting with the various apis
	clients *manager.Clients
}

type UserInfo struct {
	token          string
	LongLivedToken string
	Guild          string
	TextChannel    string
}
