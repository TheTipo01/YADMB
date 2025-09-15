package api

import (
	"net/http"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

type Api struct {
	// Server managers
	servers map[string]*manager.Server
	// Map from token to users
	tokensToUsers map[string]*discordgo.User
	// Map from userID to token
	userInfo map[string]*UserInfo
	// Bot owner
	owner map[string]struct{}
	// CLients for interacting with the various apis
	clients *manager.Clients
	// Websocket connections
	notifier *notification.Notifier
	// HTTP filesystem
	fe         http.FileSystem
	wsUpgrader *websocket.Upgrader
}

type UserInfo struct {
	token          string
	LongLivedToken string
	Guild          string
	TextChannel    string
}
