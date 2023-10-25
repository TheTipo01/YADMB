package api

import (
	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (a *Api) getQueue(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	queue := a.servers[guild].Queue.GetAllQueue()
	if len(queue) > 0 {
		queue[0].Frames = a.servers[guild].Frames
	}

	c.JSON(http.StatusOK, queue)
}

func (a *Api) addToQueue(c *gin.Context) {
	token := c.PostForm("token")
	guild := c.Param("guild")
	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	song := c.PostForm("song")
	playlist := stringToBool(c.PostForm("playlist"))
	shuffle := stringToBool(c.PostForm("shuffle"))
	loop := stringToBool(c.PostForm("loop"))
	priority := stringToBool(c.PostForm("priority"))

	i := a.interactionGenerator(u, song, playlist, shuffle, loop, priority, guild)

	switch a.servers[guild].PlayCommand(a.clients, i, playlist, a.owner) {
	case manager.Success:
		c.Status(http.StatusOK)
	case manager.NotInVC:
		c.Status(http.StatusForbidden)
	case manager.DjMode:
		c.Status(http.StatusForbidden)
	case manager.Playlist:
		c.Status(http.StatusNotAcceptable)
	}
}

func (a *Api) skip(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.ID) != nil {
		if !a.servers[guild].IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if stringToBool(c.PostForm("clean")) {
				go a.servers[guild].Clean()
				c.Status(http.StatusOK)
			} else {
				a.servers[guild].Skip <- manager.Skip
				a.servers[guild].Paused.Store(false)
				c.Status(http.StatusOK)
			}
		}
	} else {
		c.Status(http.StatusForbidden)
	}
}

func (a *Api) pause(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.ID) != nil {
		if !a.servers[guild].IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if a.servers[guild].Paused.CompareAndSwap(false, true) {
				a.servers[guild].Pause <- struct{}{}
				c.Status(http.StatusOK)
			} else {
				c.Status(http.StatusNotAcceptable)
			}
		}
	} else {
		c.Status(http.StatusForbidden)
	}
}

func (a *Api) resume(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.ID) != nil {
		if !a.servers[guild].IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if a.servers[guild].Paused.CompareAndSwap(true, false) {
				a.servers[guild].Resume <- struct{}{}
				c.Status(http.StatusOK)
			} else {
				c.Status(http.StatusNotAcceptable)
			}
		}
	} else {
		c.Status(http.StatusForbidden)
	}
}

func (a *Api) toggle(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.ID) != nil {
		if !a.servers[guild].IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if a.servers[guild].Paused.CompareAndSwap(false, true) {
				a.servers[guild].Pause <- struct{}{}
				c.Status(http.StatusOK)
			} else {
				if a.servers[guild].Paused.CompareAndSwap(true, false) {
					a.servers[guild].Resume <- struct{}{}
					c.Status(http.StatusOK)
				} else {
					c.Status(http.StatusInternalServerError)
				}
			}
		}
	} else {
		c.Status(http.StatusForbidden)
	}
}

func (a *Api) getFavorites(c *gin.Context) {
	token := c.Query("token")
	u, authorized := a.checkAuthorization(token)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, a.clients.Database.GetFavorites(u.ID))
}

func (a *Api) removeFavorite(c *gin.Context) {
	token := c.Query("token")
	u, authorized := a.checkAuthorization(token)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	name := c.Query("name")
	err := a.clients.Database.RemoveFavorite(u.ID, name)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

func (a *Api) addFavorite(c *gin.Context) {
	token := c.PostForm("token")
	u, authorized := a.checkAuthorization(token)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	name := c.PostForm("name")
	link := c.PostForm("link")
	folder := c.PostForm("folder")

	err := a.clients.Database.AddFavorite(u.ID, database.Favorite{Name: name, Link: link, Folder: folder})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

const (
	// Time allowed to write the message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func (a *Api) websocketHandler(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err == nil {
		n := make(chan notification.NotificationMessage)
		a.notifier.AddChannel(n, guild)

		pingTicker := time.NewTicker(pingPeriod)
		clean := func() {
			pingTicker.Stop()
			_ = conn.Close()
			close(n)
			a.notifier.RemoveChannel(n, guild)
		}
		defer clean()

		conn.SetCloseHandler(func(code int, text string) error {
			clean()
			return nil
		})

		// TODO: Handling websockets like this exposes the api to a DoS attack, by simpling spawning a lot of connections
		for {
			select {
			case msg, ok := <-n:
				if !ok {
					return
				}

				err = conn.WriteJSON(msg)
				if err != nil {
					return
				}
			case <-pingTicker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err = conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}
}
