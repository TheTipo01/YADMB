package manager

import (
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/bwmarrin/discordgo"
	"time"
)

// JoinVC joins the voice channel if not already joined, returns true if joined successfully
func JoinVC(i *discordgo.Interaction, channelID string, s *discordgo.Session, server *Server, isDeferred chan struct{}) bool {
	if !server.VC.IsConnected() {
		// Join the voice channel
		err := server.VC.Join(s, channelID)
		if err != nil {
			embed.SendAndDeleteEmbedInteraction(s, embed.NewEmbed().SetTitle(s.State.User.Username).AddField(constants.ErrorTitle, constants.CantJoinVC).
				SetColor(0x7289DA).MessageEmbed, i, time.Second*5, isDeferred)
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
func FindUserVoiceState(s *discordgo.Session, guildID, userID string) *discordgo.VoiceState {
	g, err := s.State.Guild(guildID)
	if err == nil {
		for _, vs := range g.VoiceStates {
			if vs.UserID == userID {
				return vs
			}
		}
	}

	return nil
}
