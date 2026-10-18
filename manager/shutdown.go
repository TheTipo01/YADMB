package manager

import (
	"context"
	"time"
)

// Shutdown stops playback, clears the queue and leaves the voice channel.
func (server *Server) Shutdown(ctx context.Context) {
	server.shuttingDown.Store(true)
	server.CancelQuit()

	// Let a Clean already started by a command finish before doing our own.
	for server.cleaning.Load() && ctx.Err() == nil {
		select {
		case <-ctx.Done():
		case <-time.After(25 * time.Millisecond):
		}
	}

	server.clean(ctx, false, true)
	server.VC.Disconnect(ctx)
}
