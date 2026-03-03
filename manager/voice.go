package manager

import (
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// JoinVC joins the voice channel if not already joined, returns true if joined successfully
func JoinVC(e *events.ApplicationCommandInteractionCreate, channelID snowflake.ID, server *Server, isDeferred chan struct{}) bool {
	if !server.VC.IsConnected() {
		// Join the voice channel
		err := server.VC.Join(channelID, e.Client())
		if err != nil {
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.CantJoinVC, false).
				SetColor(0x7289DA).Build(), e, time.Second*5, isDeferred)
			return false
		}
	}
	return true
}

// QuitVC disconnects the bot from the voice channel after 1 minute if nothing is playing
func (server *Server) QuitVC() {
	if server.Queue.IsEmpty() {
		server.VC.Disconnect()
	}
}

// FindUserVoiceState finds user current voice channel
func FindUserVoiceState(s *bot.Client, guildID, userID snowflake.ID) *discord.VoiceState {
	for vs := range s.Caches.VoiceStates(guildID) {
		if vs.UserID == userID {
			return &vs
		}
	}

	return nil
}
