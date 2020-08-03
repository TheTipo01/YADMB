package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user string) {

	if strings.Contains(link, "youtube.com") || strings.Contains(link, "youtu.be") {
		files, _ := ioutil.ReadDir("./audio_cache")

		//We check if the song 
		for _, f := range files {
			id := strings.TrimSuffix(f.Name(), ".dca")
			if strings.Contains(link, id) && f.Name() != ".dca"{
				el := Queue{"", "", id, link, user}
				queue[guildID] = append(queue[guildID], el)
				go addInfo(id, guildID)
				go playSound(s, guildID, channelID, f.Name())
				return
			}
		}
		//Gets video id
		var out []byte

		switch runtime.GOOS {
		case "linux":
			out, _ = exec.Command("youtube-dl", "--get-id", link).Output()
			break
		case "windows":
			out, _ = exec.Command("cmd", "/C", "youtube-dl", "--get-id", link).Output()
		}

		ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

		for _, id := range ids {
			link = "https://www.youtube.com/watch?v=" + id

			//Checks if video is already downloaded
			_, err := os.Stat("./audio_cache/" + id + ".dca")

			//If not, we download and convert it
			if err != nil {
				switch runtime.GOOS {
				case "linux":
					_ = exec.Command("youtube-dl", "-o", "download/"+id+".m4a", "-f 140", link).Run()
					_ = exec.Command("bash", "gen.sh", id).Run()
					break
				case "windows":
					_ = exec.Command("youtube-dl", "-o", "download/"+id+".m4a", "-f 140", link).Run()
					_ = exec.Command("gen.bat", id).Run()
				}

				err = os.Remove("./download/" + id + ".m4a")
				if err != nil {
					fmt.Println("Can't delete file", err)
				}
			}

			queue[guildID] = append(queue[guildID], Queue{"", "", id, link, user})
			go addInfo(id, guildID)
			go playSound(s, guildID, channelID, id+".dca")
		}
	}

}

func searchDownloadAndPlay(s *discordgo.Session, guildID, channelID, query, user string) {

	files, _ := ioutil.ReadDir("./audio_cache")

	for _, f := range files {

		id := strings.TrimSuffix(f.Name(), ".dca")
		if strings.Contains(query, id) && f.Name() != ".dca" {
			el := Queue{"", "", id, query, user}
			queue[guildID] = append(queue[guildID], el)
			go addInfo(id, guildID)
			go playSound(s, guildID, channelID, f.Name())
			return
		}
	}
	//Gets video id
	var out []byte

	switch runtime.GOOS {
	case "linux":
		out, _ = exec.Command("youtube-dl", "--get-id", "ytsearch:\""+query+"\"").Output()
		break
	case "windows":
		out, _ = exec.Command("cmd", "/C", "youtube-dl", "--get-id", "ytsearch:\""+query+"\"").Output()
	}

	ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	for _, id := range ids {
		link := "https://www.youtube.com/watch?v=" + id

		//Checks if video is already downloaded
		_, err := os.Stat("./audio_cache/" + id + ".dca")

		//If not, we download and convert it
		if err != nil {
			switch runtime.GOOS {
			case "linux":
				_ = exec.Command("youtube-dl", "-o", "download/"+id+".m4a", "-f 140", link).Run()
				_ = exec.Command("bash", "gen.sh", id).Run()
				break
			case "windows":
				_ = exec.Command("youtube-dl", "-o", "download/"+id+".m4a", "-f 140", link).Run()
				_ = exec.Command("gen.bat", id).Run()
			}

			err = os.Remove("./download/" + id + ".m4a")
			if err != nil {
				fmt.Println("Can't delete file", err)
			}
		}

		queue[guildID] = append(queue[guildID], Queue{"", "", id, link, user})
		go addInfo(id, guildID)
		go playSound(s, guildID, channelID, id+".dca")
	}

}

func spotifyPlaylist(s *discordgo.Session, guildID, channelID, user, playlistId string) {

	token, err := config.Token(context.Background())
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
		return
	}

	client := spotify.Authenticator{}.NewClient(token)

	playlist, err := client.GetPlaylist(spotify.ID(strings.TrimPrefix(playlistId, "spotify:playlist:")))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, track := range playlist.Tracks.Tracks {
		searchDownloadAndPlay(s, guildID, channelID, track.Track.Name + " - "+track.Track.Artists[0].Name, user)
	}

}