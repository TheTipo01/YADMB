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

func playSound(s *discordgo.Session, guildID, channelID, fileName, txtChannel string) {
	var opuslen int16

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
		if len(server[guildID].queue) == 1 {
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Cleared", "Queue cleared").SetColor(0x7289DA).MessageEmbed, txtChannel)
			server[guildID].clear = false
		}
		server[guildID].server.Unlock()
		return
	}

	// Sends now playing message
	m, err := s.ChannelMessageSendEmbed(txtChannel, NewEmbed().SetTitle(s.State.User.Username).AddField("Now playing", server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed)
	if err != nil {
		lit.Error("%s", err)
	}

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

	// Sets when we started reading file, so we known the remaining time of the song
	tmpTime := time.Now()
	server[guildID].queue[0].time = &tmpTime

	// Channel to send ok messages
	c1 := make(chan string, 1)

	for {
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

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		server[guildID].pause.Lock()
		if !server[guildID].skip {
			// Send data in a goroutine
			go func() {
				server[guildID].vc.OpusSend <- InBuf
				c1 <- "ok"
			}()

			// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
			select {
			case _ = <-c1:
				break
			case <-time.After(time.Second / 3):
				server[guildID].vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
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

	// If the song is skipped, we send a feedback message
	if server[guildID].skip && !server[guildID].clear {
		go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Skipped", server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed, txtChannel)
	}

	// Resets the skip boolean
	server[guildID].skip = false

	// Delete old message
	if m != nil {
		err = s.ChannelMessageDelete(txtChannel, m.ID)
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

func playSoundStream(s *discordgo.Session, guildID, channelID, fileName, txtChannel string, stdout io.ReadCloser) {
	var opuslen int16

	// Locks the mutex for the current server
	server[guildID].server.Lock()

	// Check if we need to clear
	if server[guildID].clear {
		removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)
		// If this is the last element, we have finished clearing the queue
		if len(server[guildID].queue) == 1 {
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Cleared", "Queue cleared").SetColor(0x7289DA).MessageEmbed, txtChannel)
			server[guildID].clear = false
		}
		server[guildID].server.Unlock()
		return
	}

	// Sends now playing message
	m, err := s.ChannelMessageSendEmbed(txtChannel, NewEmbed().SetTitle(s.State.User.Username).AddField("Now playing", server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).SetThumbnail(server[guildID].queue[0].thumbnail).MessageEmbed)
	if err != nil {
		lit.Error("%s", err)
	}

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

	// Sets when we started reading file, so we known the remaining time of the song
	tmpTime := time.Now()
	server[guildID].queue[0].time = &tmpTime

	// Channel to send ok messages
	c1 := make(chan string, 1)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(stdout, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if opuslen < 0 || err != nil {
			continue
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(stdout, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			lit.Error("Error reading from dca file: %s", err)
			break
		}

		// Stream data to discord
		server[guildID].pause.Lock()
		if !server[guildID].skip {
			// Send data in a goroutine
			go func() {
				server[guildID].vc.OpusSend <- InBuf
				c1 <- "ok"
			}()

			// So if the bot gets disconnect/moved we can rejoin the original channel and continue playing songs
			select {
			case _ = <-c1:
				break
			case <-time.After(time.Second / 3):
				server[guildID].vc, _ = s.ChannelVoiceJoin(guildID, channelID, false, true)
			}

		} else {
			server[guildID].pause.Unlock()
			break
		}
		server[guildID].pause.Unlock()
	}

	// Stop speaking
	_ = server[guildID].vc.Speaking(false)

	// If the song is skipped, we send a feedback message
	if server[guildID].skip && !server[guildID].clear {
		go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Skipped", server[guildID].queue[0].title+" - "+server[guildID].queue[0].duration+" added by "+server[guildID].queue[0].user).SetColor(0x7289DA).MessageEmbed, txtChannel)
	}

	// Resets the skip boolean
	server[guildID].skip = false

	// Delete old message
	if m != nil {
		err = s.ChannelMessageDelete(txtChannel, m.ID)
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
