package main

import (
	"context"
	"time"

	"github.com/TheTipo01/YADMB/manager"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

func initializeServer(guild snowflake.ID) *manager.Server {
	return servers.GetOrCreate(guild, &clients)
}

// getServer returns the server for the guild, creating it on demand.
func getServer(guild snowflake.ID) *manager.Server {
	return initializeServer(guild)
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

	channelID, ok := server.VC.ChannelID()
	if !ok {
		return
	}

	if countVoiceStates(server.Clients.Discord, server.GuildID, channelID) == 0 {
		ClearAndExit(server)
	}
}

func ClearAndExit(server *manager.Server) {
	server.Clean()
	server.CancelQuit()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.VC.Disconnect(ctx)
}
