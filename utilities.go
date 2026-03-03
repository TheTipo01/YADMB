package main

import (
	"time"

	"github.com/TheTipo01/YADMB/manager"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

func initializeServer(guild string) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	if _, ok := server[guild]; !ok {
		server[guild] = manager.NewServer(snowflake.MustParse(guild), &clients)
	}
}

func countVoiceStates(s *bot.Client, guild, channel snowflake.ID) int {
	var count int

	for vs := range s.Caches.VoiceStates(guild) {
		if vs.ChannelID != nil && *vs.ChannelID == channel && vs.UserID != s.ApplicationID {
			count++
		}
	}

	return count
}

// QuitIfEmptyVoiceChannel stops the music if the bot is alone in the voice channel
func QuitIfEmptyVoiceChannel(server *manager.Server) {
	time.Sleep(1 * time.Minute)

	if server.VC.IsConnected() && countVoiceStates(server.Clients.Discord, snowflake.MustParse(server.GuildID), snowflake.MustParse(server.VC.GetChannelID().String())) == 0 {
		ClearAndExit(server)
	}
}

func ClearAndExit(server *manager.Server) {
	server.Clean()
	server.VC.Disconnect()
}
