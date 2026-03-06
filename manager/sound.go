package manager

import (
	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/queue"
)

// Plays a song in DCA format
func (server *Server) playSound(el *queue.Element) (SkipReason, error) {
	go notify(notification.NotificationMessage{Notification: notification.Playing, Guild: server.GuildID})

	provider := newDCAFrameProvider(el, server)
	server.VC.SetOpusFrameProvider(provider)

	select {
	case skipReason := <-server.Skip:
		provider.Close()
		return skipReason, nil
	case <-provider.done:
		return Finished, nil
	}
}
