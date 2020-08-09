package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

//Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user, txtChannel string) {

	files, _ := ioutil.ReadDir("./audio_cache")

	//We check if the song is already downloaded
	for _, f := range files {
		id := strings.TrimSuffix(f.Name(), ".dca")
		if strings.Contains(link, id) && f.Name() != ".dca" {
			el := Queue{"", "", id, link, user}
			queue[guildID] = append(queue[guildID], el)
			addInfo(id, guildID)
			go playSound(s, guildID, channelID, f.Name(), txtChannel, findQueuePointer(guildID, id))
			return
		}
	}

	if strings.Contains(link, "youtube.com") || strings.Contains(link, "youtu.be") {

		//Gets info about songs
		out, _ := exec.Command("youtube-dl", "--get-id", "-e", "--get-duration", link).Output()

		//Parse output as string, splitting it on every newline
		strOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

		//We generate a temporary temp queue, parsing info from youtube-dl
		tmpQueue := make([]Queue, (len(strOut)+1)/3)
		j := 0
		for i := 0; i < len(strOut); i += 3 {
			tmpQueue[j].title = strOut[i]
			tmpQueue[j].id = strOut[i+1]
			tmpQueue[j].duration = strOut[i+2]
			j++
		}

		for _, el := range tmpQueue {
			link = "https://www.youtube.com/watch?v=" + el.id

			//We only send enqueued message if it's a single song
			if len(tmpQueue) == 1 {
				go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", link).SetColor(0x7289DA).MessageEmbed, txtChannel)
			}

			//Checks if video is already downloaded
			_, err := os.Stat("./audio_cache/" + el.id + ".dca")

			//If not, we download and convert it
			if err != nil {
				_ = exec.Command("youtube-dl", "-o", "download/"+el.id+".m4a", "-f 140", link).Run()

				switch runtime.GOOS {
				case "linux":
					_ = exec.Command("bash", "gen.sh", el.id).Run()
					break
				case "windows":
					_ = exec.Command("gen.bat", el.id).Run()
				}

				err = os.Remove("./download/" + el.id + ".m4a")
				if err != nil {
					fmt.Println("Can't delete file", err)
				}
			}

			el := Queue{el.title, el.duration, el.id, link, user}

			queue[guildID] = append(queue[guildID], el)
			go playSound(s, guildID, channelID, el.id+".dca", txtChannel, findQueuePointer(guildID, el.id))
		}

		//If it's a playlist, we send a final message telling the users that we enqueued all the song
		if len(tmpQueue) != 1 {
			go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", strconv.Itoa(len(tmpQueue)+1)+" songs").SetColor(0x7289DA).MessageEmbed, txtChannel)
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
