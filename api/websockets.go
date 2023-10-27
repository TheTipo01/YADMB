package api

import (
	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync/atomic"
	"time"
)

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

	conn, err := a.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err == nil {
		n := make(chan notification.NotificationMessage)
		a.notifier.AddChannel(n, guild)

		pingTicker := time.NewTicker(pingPeriod)

		// TODO: while I do feel like a gigabrain for coming up with this solution, I'm not sure if this is the best way to do this
		counter := atomic.Bool{}
		clean := func() {
			if counter.CompareAndSwap(false, true) {
				pingTicker.Stop()
				_ = conn.Close()
				close(n)
				a.notifier.RemoveChannel(n, guild)
			}
		}

		conn.SetCloseHandler(func(code int, text string) error {
			clean()
			return nil
		})

		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(pongWait))
		})

		// Writer
		go func() {
			defer clean()

			for {
				select {
				case msg, ok := <-n:
					if !ok {
						return
					}

					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
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

		}()

		// Pong reader
		go func() {
			defer clean()
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()
	}
}

func (a *Api) HandleNotifications() {
	for {
		select {
		case n := <-manager.Notifications:
			a.notifier.Notify(n.Guild, n)
		}
	}
}
