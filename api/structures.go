package api

import (
	"sync"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/disgoorg/snowflake/v2"
)

type Api struct {
	// Server managers
	servers *manager.Servers
	// Map from token to users
	tokensToUsers sync.Map
	// Map from userID to token
	userInfo sync.Map
	// Serializes compound updates to the two maps above
	userMu sync.Mutex
	// Bot owner
	owner map[snowflake.ID]struct{}
	// CLients for interacting with the various apis
	clients *manager.Clients
	// Websocket connections
	notifier *notification.Notifier
	// Origin for WebSocket CORS
	origin string
}

type UserInfo struct {
	token          string
	LongLivedToken string
	Guild          snowflake.ID
	TextChannel    snowflake.ID
}
