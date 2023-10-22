package api

import (
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (a *Api) getQueue(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, a.servers[guild].Queue.GetAllQueue())
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

	a.servers[guild].PlayCommand(a.clients, i, playlist, a.owner)

	c.Status(http.StatusOK)
}

func (a *Api) skip(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	a.servers[guild].Skip <- manager.Skip
	a.servers[guild].Paused.Store(false)
	c.Status(http.StatusOK)
}

func (a *Api) pause(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if a.servers[guild].Paused.CompareAndSwap(false, true) {
		a.servers[guild].Pause <- struct{}{}
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotAcceptable)
	}
}

func (a *Api) resume(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if a.servers[guild].Paused.CompareAndSwap(true, false) {
		a.servers[guild].Resume <- struct{}{}
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotAcceptable)
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
	token := c.PostForm("token")
	u, authorized := a.checkAuthorization(token)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	name := c.PostForm("name")
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

func (a *Api) toggle(c *gin.Context) {
	token := c.Query("token")
	guild := c.Param("guild")
	_, authorized := a.checkAuthorizationAndGuild(token, guild)
	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

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
