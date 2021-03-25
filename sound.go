package main

import (
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os"
	"strings"
	"time"
)

// Plays a song from a given file
func playSound(s *discordgo.Session, guildID, channelID, fileName string, i *discordgo.Interaction) {
	var (
		opuslen int16
		skip    bool
	)

	// Locks the mutex for the current server
	server[guildID].server.Lock()

	file, err := os.Open("./audio_cache/" + fileName)
	if err != nil {
		lit.Error("Error opening dca file: %s", err)
		server[guildID].server.Unlock()
		return
	}

	// Check if we need to clear
	if server[guildID].clear {
		removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)
		// If this is the last element, we have finished clearing the queue
		if len(server[guildID].queue) == 0 {
			server[guildID].clear = false
		}
		server[guildID].server.Unlock()
		return
	}

	// Sends now playing message
	modfyInteraction(s, NewEmbed().SetTitle(s.State.User.Username).
		AddField("Now playing", server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+
			" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).
		SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed, i)

	// Join the provided voice channel.
	if server[guildID].vc == nil || server[guildID].vc.ChannelID != channelID {
		server[guildID].vc, err = s.ChannelVoiceJoin(guildID, channelID, false, true)
		if err != nil {
			lit.Error("%s", err)
			server[guildID].server.Unlock()
			return
		}
	}

	// Start speaking.
	_ = server[guildID].vc.Speaking(true)
	server[guildID].skip = false

	for {
		if server[guildID].queue[0].segments[server[guildID].queue[0].frame] {
			skip = !skip
		}

		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Keep count of the frames of the song
		server[guildID].queue[0].frame++

		if skip {
			continue
		}

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		server[guildID].pause.Lock()
		if !server[guildID].skip {
			select {
			case server[guildID].vc.OpusSend <- InBuf:
				break
			case <-time.After(time.Second / 3):
				server[guildID].vc, _ = s.ChannelVoiceJoin(guildID, server[guildID].queue[0].channel, false, true)
			}

		} else {
			server[guildID].pause.Unlock()
			break
		}
		server[guildID].pause.Unlock()
	}

	// Close the file
	_ = file.Close()

	// Stop speaking
	_ = server[guildID].vc.Speaking(false)

	// Resets the skip boolean
	server[guildID].skip = false

	// Remove from queue the song
	removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)

	// If this is the last song, we wait a minute before disconnecting from the voice channel
	if len(server[guildID].queue) == 0 {
		go quitVC(s, i)
	}

	// Releases the mutex lock for the server
	server[guildID].server.Unlock()

}

// Plays a song from a io.Reader
func soundStream(s *discordgo.Session, guildID, channelID, fileName string, i *discordgo.Interaction, stdout io.Reader) {
	var (
		opuslen int16
		skip    bool
	)

	// Locks the mutex for the current server
	server[guildID].server.Lock()

	// Check if we need to clear
	if server[guildID].clear {
		removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)
		// If this is the last element, we have finished clearing the queue
		if len(server[guildID].queue) == 0 {
			server[guildID].clear = false
		}
		server[guildID].server.Unlock()
		return
	}

	// Sends now playing message
	modfyInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Now playing",
		server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed,
		i)

	var err error
	// Join the provided voice channel.
	if server[guildID].vc == nil || server[guildID].vc.ChannelID != channelID {
		server[guildID].vc, err = s.ChannelVoiceJoin(guildID, channelID, false, true)
		if err != nil {
			lit.Error("%s", err)
			server[guildID].server.Unlock()
			return
		}
	}

	// Start speaking.
	_ = server[guildID].vc.Speaking(true)
	server[guildID].skip = false

	for {
		if server[guildID].queue[0].segments[server[guildID].queue[0].frame] {
			skip = !skip
		}

		// Read opus frame length from dca file.
		err = binary.Read(stdout, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(stdout, binary.LittleEndian, &InBuf)

		// Keep count of the frames of the song
		server[guildID].queue[0].frame++

		if skip {
			continue
		}

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error streaming from dca file: %s", err)
			break
		}

		// Stream data to discord
		server[guildID].pause.Lock()
		if !server[guildID].skip {
			select {
			case server[guildID].vc.OpusSend <- InBuf:
				break
			case <-time.After(time.Second / 3):
				server[guildID].vc, _ = s.ChannelVoiceJoin(guildID, server[guildID].queue[0].channel, false, true)
			}

		} else {
			server[guildID].pause.Unlock()
			break
		}
		server[guildID].pause.Unlock()
	}

	// Stop speaking
	_ = server[guildID].vc.Speaking(false)

	// Resets the skip boolean
	server[guildID].skip = false

	// Remove from queue the song
	removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)

	// If this is the last song, we wait a minute before disconnecting from the voice channel
	if len(server[guildID].queue) == 0 {
		go quitVC(s, i)
	}

	// Releases the mutex lock for the server
	server[guildID].server.Unlock()

}
