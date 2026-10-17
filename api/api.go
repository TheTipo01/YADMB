package api

import (
	"embed"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/dchest/uniuri"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func NewApi(servers *manager.Servers, address string, owner map[snowflake.ID]struct{}, clients *manager.Clients, buildFS *embed.FS, origin string) *Api {
	r := gin.New()

	conf := cors.DefaultConfig()
	conf.AllowMethods = []string{"GET", "POST", "DELETE"}
	conf.AllowOrigins = []string{origin}

	r.Use(gin.Recovery(), cors.New(conf))

	a := Api{
		servers:  servers,
		owner:    owner,
		clients:  clients,
		notifier: notification.NewNotifier(),
		origin:   origin,
	}

	r.GET("/queue/:guild", a.getQueue)
	r.POST("/queue/:guild", a.addToQueue)
	r.DELETE("/queue/:guild", a.skip)
	r.GET("/song/resume/:guild", a.resume)
	r.GET("/song/pause/:guild", a.pause)
	r.GET("/song/toggle/:guild", a.toggle)
	r.GET("/favorites", a.getFavorites)
	r.POST("/favorites", a.addFavorite)
	r.DELETE("/favorites", a.removeFavorite)
	r.GET("/ws/:guild", a.websocketHandler)
	r.Use(static.Serve("/", EmbedFolder(buildFS, "web/build")))
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	if strings.HasPrefix(address, "unix://") {
		socket := strings.TrimPrefix(address, "unix://")
		err := os.Remove(socket)
		if err != nil {
			slog.Error("Failed to remove existing socket", "error", err)
		}

		go r.RunUnix(socket)

		err = os.Chmod(socket, 0777)
		if err != nil {
			slog.Error("Failed to change permission for socket", "error", err)
		}
	} else {
		go r.Run(address)
	}

	return &a
}

// AddUser adds a user to the api, returning the token.
// If the user is already in the api, it removes it and generates a new one.
func (a *Api) AddUser(user *discord.Member, userInfo UserInfo) string {
	a.userMu.Lock()
	defer a.userMu.Unlock()

	userID := user.User.ID

	if value, ok := a.userInfo.Load(userID); ok {
		old := value.(*UserInfo)
		if old.token != "" {
			a.tokensToUsers.Delete(old.token)
		}

		if old.LongLivedToken == "" {
			a.userInfo.Delete(userID)
		}
	}

	// Generate a new token until it is unique
	var token string
	for {
		token = uniuri.NewLen(32)
		if _, ok := a.tokensToUsers.Load(token); !ok {
			break
		}
	}

	a.tokensToUsers.Store(token, user)

	if value, ok := a.userInfo.Load(userID); ok {
		userInfo.LongLivedToken = value.(*UserInfo).LongLivedToken
	}
	userInfo.token = token

	a.userInfo.Store(userID, &userInfo)

	return token
}

func (a *Api) AddLongLivedToken(user *discord.Member, userInfo UserInfo) {
	a.userMu.Lock()
	defer a.userMu.Unlock()

	a.tokensToUsers.Store(userInfo.LongLivedToken, user)

	userID := user.User.ID
	if value, ok := a.userInfo.Load(userID); ok {
		userInfo.token = value.(*UserInfo).token
	}
	a.userInfo.Store(userID, &userInfo)
}
