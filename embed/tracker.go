package embed

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type messageKey struct {
	channelID snowflake.ID
	messageID snowflake.ID
}

type interactionKey struct {
	applicationID snowflake.ID
	token         string
}

var (
	trackerMu sync.Mutex

	trackedMessages     = make(map[messageKey]struct{})
	trackedInteractions = make(map[interactionKey]struct{})

	shuttingDown atomic.Bool
)

// BeginShutdown makes future tracked messages delete themselves immediately.
func BeginShutdown() {
	shuttingDown.Store(true)
}

// TrackMessage records a message so it can be removed during shutdown.
// It returns false when shutdown has already started.
func TrackMessage(message *discord.Message) bool {
	if message == nil {
		return true
	}

	if shuttingDown.Load() {
		return false
	}

	trackerMu.Lock()
	defer trackerMu.Unlock()

	if shuttingDown.Load() {
		return false
	}

	trackedMessages[messageKey{channelID: message.ChannelID, messageID: message.ID}] = struct{}{}
	return true
}

// UntrackMessage removes a message from the shutdown tracker.
func UntrackMessage(channelID, messageID snowflake.ID) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	delete(trackedMessages, messageKey{channelID: channelID, messageID: messageID})
}

// TrackInteraction records an interaction response so it can be removed during shutdown.
// It returns false when shutdown has already started.
func TrackInteraction(applicationID snowflake.ID, token string) bool {
	if token == "" {
		return true
	}

	if shuttingDown.Load() {
		return false
	}

	trackerMu.Lock()
	defer trackerMu.Unlock()

	if shuttingDown.Load() {
		return false
	}

	trackedInteractions[interactionKey{applicationID: applicationID, token: token}] = struct{}{}
	return true
}

// UntrackInteraction removes an interaction response from the shutdown tracker.
func UntrackInteraction(applicationID snowflake.ID, token string) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	delete(trackedInteractions, interactionKey{applicationID: applicationID, token: token})
}

// DeleteAll deletes every message and interaction response tracked so far.
func DeleteAll(ctx context.Context, client *bot.Client) {
	if client == nil {
		return
	}

	BeginShutdown()

	trackerMu.Lock()
	messages := trackedMessages
	interactions := trackedInteractions

	trackedMessages = make(map[messageKey]struct{})
	trackedInteractions = make(map[interactionKey]struct{})
	trackerMu.Unlock()

	for key := range messages {
		if ctx.Err() != nil {
			break
		}

		if err := client.Rest.DeleteMessage(key.channelID, key.messageID, rest.WithCtx(ctx)); err != nil {
			slog.Debug("Could not delete message during shutdown", "channel", key.channelID, "message", key.messageID, "error", err)
		}
	}

	for key := range interactions {
		if ctx.Err() != nil {
			break
		}

		if err := client.Rest.DeleteInteractionResponse(key.applicationID, key.token, rest.WithCtx(ctx)); err != nil {
			slog.Debug("Could not delete interaction response during shutdown", "application", key.applicationID, "error", err)
		}
	}
}
