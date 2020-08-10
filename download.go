package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user, txtChannel string) {

	files, _ := ioutil.ReadDir("./audio_cache")

	//We check if the song is already downloaded
	for _, f := range files {
		id := strings.TrimSuffix(f.Name(), ".dca")
		if strings.Contains(link, id) && f.Name() != ".dca" {
			el := Queue{"", "", id, link, user, nil, nil, 0}
			queue[guildID] = append(queue[guildID], el)
			addInfo(id, guildID)
			go playSound(s, guildID, channelID, f.Name(), txtChannel, findQueuePointer(guildID, id))
			return
		}
	}

	go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", link).SetColor(0x7289DA).MessageEmbed, txtChannel)

	//Gets info about songs
	out, _ := exec.Command("youtube-dl", "--ignore-errors", "-q", "--no-warnings", "-j", link).Output()

	//Parse output as string, splitting it on every newline
	strOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	var ytdl YoutubeDL
	//We parse every track as individual json, because youtube-dl
	for _, sos := range strOut {
		_ = json.Unmarshal([]byte(sos), &ytdl)

		//We search through the formats
		for _, f := range ytdl.Formats {
			//If the protocol is either http or https, and the format is audio and bitrate > 128, we download the song
			if f.Protocol == "http" || f.Protocol == "https" {
				if strings.Contains(f.Format, "audio only") && f.Abr >= 128 {
					//Checks if video is already downloaded
					_, err := os.Stat("./audio_cache/" + ytdl.ID + ".dca")

					//If not, we download and convert it
					if err != nil {
						err := downloadFile("./download/"+ytdl.ID+"."+f.Ext, f.URL)
						if err != nil {
							fmt.Println(err)
							break
						}

						switch runtime.GOOS {
						case "linux":
							_ = exec.Command("bash", "gen.sh", ytdl.ID, ytdl.ID+"."+f.Ext).Run()
							break
						case "windows":
							_ = exec.Command("gen.bat", ytdl.ID, ytdl.ID+"."+f.Ext).Run()
						}

						err = os.Remove("./download/" + ytdl.ID + "." + f.Ext)
					}

					el := Queue{ytdl.Title, formatDuration(ytdl.Duration), ytdl.ID, ytdl.WebpageURL, user, nil, nil, 0}

					queue[guildID] = append(queue[guildID], el)
					go playSound(s, guildID, channelID, el.id+".dca", txtChannel, findQueuePointer(guildID, ytdl.ID))

					break
				}
			}
		}
	}

}

//Searches a song from the query on youtube
func searchDownloadAndPlay(s *discordgo.Session, guildID, channelID, query, user, txtChannel string) {
	//Gets video id
	out, _ := exec.Command("youtube-dl", "--get-id", "ytsearch:\""+query+"\"").Output()
	ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	for _, id := range ids {
		downloadAndPlay(s, guildID, channelID, "https://www.youtube.com/watch?v="+id, user, txtChannel)
	}

}

//Enqueues song from a spotify playlist, searching them on youtube
func spotifyPlaylist(s *discordgo.Session, guildID, channelID, user, playlistId, txtChannel string) {

	playlist, err := client.GetPlaylist(spotify.ID(strings.TrimPrefix(playlistId, "spotify:playlist:")))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, track := range playlist.Tracks.Tracks {
		go searchDownloadAndPlay(s, guildID, channelID, track.Track.Name+" - "+track.Track.Artists[0].Name, user, txtChannel)
	}

}
