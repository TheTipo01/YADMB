package manager

import "context"

// Shutdown stops playback, clears the queue and leaves the voice channel.
func (server *Server) Shutdown(ctx context.Context) {
	server.shuttingDown.Store(true)
	server.CancelQuit()
	server.clean(ctx, false, true)
	server.VC.Disconnect(ctx)
}
