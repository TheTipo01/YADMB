package api

import (
	"net/http"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gorilla/websocket"
)

type Api struct {
	// Server managers
	servers map[snowflake.ID]*manager.Server
	// Map from token to users
	tokensToUsers map[string]*discord.Member
	// Map from userID to token
	userInfo map[snowflake.ID]*UserInfo
	// Bot owner
	owner map[snowflake.ID]struct{}
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
	Guild          snowflake.ID
	TextChannel    snowflake.ID
}
