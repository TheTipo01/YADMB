package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Plays a song from a io.Reader if specified, or tries to open a file with the given fileName
func playSound(s *discordgo.Session, guildID, channelID, fileName string, i *discordgo.Interaction, in io.Reader, c *chan int, cmds []*exec.Cmd) {
	if c != nil {
		go deleteInteraction(s, i, c)
	}

	// Locks the mutex for the current server
	server[guildID].server.Lock()
	if len(cmds) > 0 {
		server[guildID].stream.Lock()
		cmdsStart(cmds)
	}

	var (
		opuslen int16
		skip    bool
		file    *os.File
		err     error
	)

	if in == nil {
		file, err = os.Open("./audio_cache/" + fileName)
		if err != nil {
			lit.Error("Error opening dca file: %s", err)
			server[guildID].server.Unlock()
			return
		}

		in = file
	}

	// Check if we need to clear
	if server[guildID].clear {
		removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)
		// If this is the last element, we have finished clearing the queue
		if len(server[guildID].queue) == 0 {
			err = s.InteractionResponseDelete(s.State.User.ID, i)
			if err != nil {
				lit.Error("InteractionResponseDelete: %s", err.Error())
			}

			server[guildID].clear = false
		}
		server[guildID].server.Unlock()
		return
	}

	// Sends now playing message
	m := sendEmbed(s, NewEmbed().SetTitle(s.State.User.Username).
		AddField("Now playing", fmt.Sprintf("[%s](%s) - %s added by %s", server[guildID].queue[0].title,
			server[guildID].queue[0].link, server[guildID].queue[0].duration, server[guildID].queue[0].user)).
		SetColor(0x7289DA).SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed, i.ChannelID)

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
		err = binary.Read(in, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(in, binary.LittleEndian, &InBuf)

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

	// If we are using a file, close it
	if file != nil {
		_ = file.Close()
	}

	// Stop speaking
	_ = server[guildID].vc.Speaking(false)

	// Resets the skip boolean
	server[guildID].skip = false

	// Delete old message
	if m != nil {
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			lit.Error("%s", err)
		}

		deleteMessages(s, server[guildID].queue[0].messageID)
	}

	// Remove from queue the song
	removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)

	// If this is the last song, we wait a minute before disconnecting from the voice channel
	if len(server[guildID].queue) == 0 {
		go quitVC(guildID)
	}

	// Releases the mutex lock for the server
	server[guildID].server.Unlock()
}
