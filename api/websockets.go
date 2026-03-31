package api

import (
	"context"
	"net/http"
	"time"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/manager"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gin-gonic/gin"
)

const (
	// Time allowed to write the message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
)

func (a *Api) websocketHandler(c *gin.Context) {
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

	// Check origin before accepting the WebSocket connection
	if !a.checkOrigin(c.Request) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Accept the WebSocket connection
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "")

	// Create a context for the connection lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := make(chan notification.NotificationMessage, 10)
	a.notifier.AddChannel(n, guild)
	defer a.notifier.RemoveChannel(n, guild)

	// Channel to signal when goroutines are done
	done := make(chan struct{}, 2)

	// Ping ticker
	pingTicker := time.NewTicker((pongWait * 9) / 10)
	defer pingTicker.Stop()

	// Writer goroutine - sends messages and pings to client
	go func() {
		defer func() { done <- struct{}{} }()

		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-n:
				if !ok {
					return
				}

				writeCtx, cancel := context.WithTimeout(ctx, writeWait)
				if err := wsjson.Write(writeCtx, conn, msg); err != nil {
					cancel()
					return
				}
				cancel()

			case <-pingTicker.C:
				writeCtx, cancel := context.WithTimeout(ctx, writeWait)
				if err := conn.Ping(writeCtx); err != nil {
					cancel()
					return
				}
				cancel()
			}
		}
	}()

	// Reader goroutine - keeps connection alive by reading pong messages
	go func() {
		defer func() { done <- struct{}{} }()

		for {
			select {
			case <-ctx.Done():
				return

			default:
				readCtx, cancel := context.WithTimeout(ctx, pongWait)
				_, _, err := conn.Read(readCtx)
				cancel()

				if err != nil {
					// Connection closed or timeout
					return
				}
			}
		}
	}()

	// Wait for first goroutine to exit, then stop the other, then wait it too
	<-done
	cancel()
	<-done
}

func (a *Api) HandleNotifications() {
	for {
		select {
		case n := <-manager.Notifications:
			a.notifier.Notify(n.Guild, n)
		}
	}
}
