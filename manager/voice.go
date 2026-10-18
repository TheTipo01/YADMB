package manager

import (
	"context"
	"errors"
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/vc"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// EnsureVoice makes sure the bot is connected to channelID.
// If the bot is idle in another channel it moves there; if it is playing in
// another channel it returns vc.ErrAlreadyConnected.
func (server *Server) EnsureVoice(channelID snowflake.ID) error {
	if server.shuttingDown.Load() {
		return errors.New("server is shutting down")
	}

	if server.VC.IsConnected() {
		current, ok := server.VC.ChannelID()
		if ok && current == channelID {
			return nil
		}

		if server.IsPlaying() {
			return vc.ErrAlreadyConnected
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		server.VC.Disconnect(ctx)
		cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return server.VC.Join(ctx, channelID, server.Clients.Discord.VoiceManager)
}

// JoinVC joins the voice channel if needed, returns true if joined successfully.
func JoinVC(e *events.ApplicationCommandInteractionCreate, channelID snowflake.ID, server *Server, isDeferred chan struct{}) bool {
	if err := server.EnsureVoice(channelID); err != nil {
		msg := constants.CantJoinVC
		if errors.Is(err, vc.ErrAlreadyConnected) {
			msg = constants.BotInAnotherVC
		}

		embed.SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(BotName).AddField(constants.ErrorTitle, msg, false).
			WithColor(0x7289DA), e, time.Second*5, isDeferred)
		return false
	}

	return true
}

// FindUserVoiceState finds user current voice channel
func FindUserVoiceState(s *bot.Client, guildID, userID snowflake.ID) *discord.VoiceState {
	v, found := s.Caches.VoiceState(guildID, userID)

	if !found {
		return nil
	}

	return &v
}
