package main

import (
	"github.com/TheTipo01/YADMB/manager"
	"github.com/bwmarrin/discordgo"
	"time"
)

func initializeServer(guild string) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	if _, ok := server[guild]; !ok {
		server[guild] = manager.NewServer(guild, &clients)
	}
}

func countVoiceStates(s *discordgo.Session, guild, channel string) (count int) {
	g, err := s.State.Guild(guild)
	if err == nil {
		s.RLock()
		defer s.RUnlock()

		for _, vs := range g.VoiceStates {
			if vs.ChannelID == channel && vs.UserID != s.State.User.ID {
				count++
			}
		}
	}

	return
}

// QuitIfEmptyVoiceChannel stops the music if the bot is alone in the voice channel
func QuitIfEmptyVoiceChannel(server *manager.Server) {
	time.Sleep(1 * time.Minute)

	if server.VC.IsConnected() && countVoiceStates(server.Clients.Discord, server.GuildID, server.VC.GetChannelID()) == 0 {
		ClearAndExit(server)
	}
}

func ClearAndExit(server *manager.Server) {
	server.Clean()
	server.VC.Disconnect()
}
