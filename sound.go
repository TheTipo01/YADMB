package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"strings"
	"time"
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

	//Check if we need to clear
	if clear[guildID] {
		removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)
		//If this is the last element, we have finished clearing the queue
		if len(queue[guildID]) == 1 {
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Cleared", "Queue cleared").SetColor(0x7289DA).MessageEmbed, txtChannel)
			clear[guildID] = false
		}
		server[guildID].Unlock()
		return
	}

	//Sends now playing message
	m, err := s.ChannelMessageSendEmbed(txtChannel, NewEmbed().SetTitle(s.State.User.Username).AddField("Now playing", queue[guildID][0].title+" - "+queue[guildID][0].duration+" added by "+queue[guildID][0].user).SetColor(0x7289DA).MessageEmbed)
	if err != nil {
		fmt.Println(err)
	}

	//Join the provided voice channel.
	if vc[guildID] == nil || vc[guildID].ChannelID != channelID {
		vc[guildID], err = s.ChannelVoiceJoin(guildID, channelID, false, true)
		if err != nil {
			return
		}
	}

	//Start speaking.
	_ = vc[guildID].Speaking(true)
	skip[guildID] = false

	//Sets when we started reading file, so we known the remaining time of the song
	tmpTime := time.Now()
	queue[guildID][0].time = &tmpTime

	for {
		//Read opus frame length from dca file.
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

		//Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		//Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			break
		}

		//Stream data to discord
		pause[guildID].Lock()
		if !skip[guildID] {
			vc[guildID].OpusSend <- InBuf
		} else {
			pause[guildID].Unlock()
			break
		}
		pause[guildID].Unlock()
	}

	//Stop speaking
	_ = vc[guildID].Speaking(false)

	//If the song is skipped, we send a feedback message
	if skip[guildID] && !clear[guildID] {
		go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Skipped", queue[guildID][0].title+" - "+queue[guildID][0].duration+" added by "+queue[guildID][0].user).SetColor(0x7289DA).MessageEmbed, txtChannel)
	}

	//Resets the skip boolean
	skip[guildID] = false

	//Delete old message
	err = s.ChannelMessageDelete(txtChannel, m.ID)
	if err != nil {
		fmt.Println(err)
	}

	//Remove from queue the song
	removeFromQueue(strings.TrimSuffix(fileName, ".dca"), guildID)

	//Releases the mutex lock for the server
	server[guildID].Unlock()

}
