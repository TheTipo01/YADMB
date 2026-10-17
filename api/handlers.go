package api

import (
	"net/http"
	"strings"

	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
)

func (a *Api) getQueue(c *gin.Context) {
	token := c.Query("token")

	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	queue := srv.Queue.GetAllQueue()
	if len(queue) > 0 {
		queue[0].Frames = int(srv.Frames.Load())
		isPaused := srv.Paused.Load()
		queue[0].IsPaused = &isPaused
	}

	c.JSON(http.StatusOK, queue)
}

func (a *Api) addToQueue(c *gin.Context) {
	token := c.PostForm("token")
	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	song := strings.TrimSpace(c.PostForm("song"))
	playlist := stringToBool(c.PostForm("playlist"))
	shuffle := stringToBool(c.PostForm("shuffle"))
	loop := stringToBool(c.PostForm("loop"))
	priority := stringToBool(c.PostForm("priority"))

	status := manager.NotInVC

	if srv.DjModeCheck(*u, a.owner) {
		status = manager.DjMode
	} else {
		if !playlist {
			song, err = manager.FilterPlaylist(song)
		}

		if err != nil {
			status = manager.Playlist
		} else if vs := manager.FindUserVoiceState(a.clients.Discord, guild, u.User.ID); vs == nil {
			status = manager.NotInVC
		} else if err := srv.EnsureVoice(*vs.ChannelID); err != nil {
			status = manager.NotInVC
		} else {
			userInfo, ok := a.getUserInfo(u.User.ID)
			if !ok {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			srv.Play(manager.PlayEvent{
				Username:    u.User.Username,
				Song:        song,
				Clients:     a.clients,
				Event:       nil,
				Random:      shuffle,
				Loop:        loop,
				Priority:    priority,
				IsDeferred:  nil,
				TextChannel: userInfo.TextChannel,
			})

			status = manager.Success
		}
	}

	switch status {
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
	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.User.ID) != nil {
		if !srv.IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if stringToBool(c.Query("clean")) {
				go srv.Clean()
				c.Status(http.StatusOK)
			} else {
				srv.SkipSong(manager.Skip)
				srv.Paused.Store(false)
				c.Status(http.StatusOK)
			}
		}
	} else {
		c.Status(http.StatusForbidden)
	}
}

func (a *Api) pause(c *gin.Context) {
	token := c.Query("token")
	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.User.ID) != nil {
		if !srv.IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if srv.PausePlayback() {
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
	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.User.ID) != nil {
		if !srv.IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else {
			if srv.ResumePlayback() {
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
	guild, err := snowflake.Parse(c.Param("guild"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	srv, ok := a.servers.Get(guild)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if manager.FindUserVoiceState(a.clients.Discord, guild, u.User.ID) != nil {
		if !srv.IsPlaying() {
			c.Status(http.StatusNotAcceptable)
		} else if srv.Paused.Load() {
			if srv.ResumePlayback() {
				c.Status(http.StatusOK)
			} else {
				c.Status(http.StatusInternalServerError)
			}
		} else {
			if srv.PausePlayback() {
				c.Status(http.StatusOK)
			} else {
				c.Status(http.StatusInternalServerError)
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

	c.JSON(http.StatusOK, a.clients.Database.GetFavorites(u.User.ID))
}

func (a *Api) removeFavorite(c *gin.Context) {
	token := c.Query("token")
	u, authorized := a.checkAuthorization(token)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	name := c.Query("name")
	err := a.clients.Database.RemoveFavorite(u.User.ID, name)
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

	err := a.clients.Database.AddFavorite(u.User.ID, database.Favorite{Name: name, Link: link, Folder: folder})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}
