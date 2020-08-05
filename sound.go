package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"strings"
)

func playSound(s *discordgo.Session, guildID, channelID, fileName, txtChannel string) {
	var opuslen int16

	file, err := os.Open("./audio_cache/" + fileName)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return
	}

	//Locks the mutex for the current server
	server[guildID].Lock()

	// Join the provided voice channel.
	if vc[guildID] == nil || vc[guildID].ChannelID != channelID {
		vc[guildID], err = s.ChannelVoiceJoin(guildID, channelID, false, false)
		if err != nil {
			return
		}
	}

	// Start speaking.
	_ = vc[guildID].Speaking(true)
	skip[guildID] = true

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				break
			}
			break
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		// Stream data to discord
		if skip[guildID] {
			vc[guildID].OpusSend <- InBuf
		} else {
			break
		}
	}

	//Resets the skip boolean
	skip[guildID] = true

	// Stop speaking
	_ = vc[guildID].Speaking(false)

	// Releases the mutex lock for the server
	server[guildID].Unlock()

	//Remove from queue the song
	removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID, txtChannel, s)

}
