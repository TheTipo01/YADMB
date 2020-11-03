package main

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/zmb3/spotify"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user, txtChannel string, random bool) {
	go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", link).SetColor(0x7289DA).MessageEmbed, txtChannel)

	// Check if the song is the db, to speedup things
	el := checkInDb(link)
	if el.title != "" {
		info, err := os.Stat("./audio_cache/" + el.id + ".dca")
		if err == nil && info.Size() > 0 {
			el.user = user
			queue[guildID] = append(queue[guildID], el)
			go playSound(s, guildID, channelID, el.id+".dca", txtChannel)
			return
		}
	}

	// Gets info about songs
	out, err := exec.Command("youtube-dl", "--ignore-errors", "-q", "--no-warnings", "-j", link).Output()
	if err != nil {
		lit.Error("Can't get info about song: %s", err)
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about song!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, txtChannel)
		return
	}

	// Parse output as string, splitting it on every newline
	strOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	var ytdl YoutubeDL

	// If we want to play the song in a random order, we just shuffle the slice
	if random {
		strOut = shuffle(strOut)
	}

	// We parse every track as individual json, because youtube-dl
	for _, singleJSON := range strOut {
		_ = json.Unmarshal([]byte(singleJSON), &ytdl)
		fileName := ytdl.ID + "-" + ytdl.Extractor
		el := Queue{ytdl.Title, formatDuration(ytdl.Duration), fileName, ytdl.WebpageURL, user, nil, 0, "", nil}

		// Checks if video is already downloaded
		info, err := os.Stat("./audio_cache/" + fileName + ".dca")

		// We add the song to the db, for faster parsing
		addToDb(el)

		// If we have a single song, we also add it with the given link
		if len(strOut) == 1 {
			el.link = link
			addToDb(el)
		}

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			// Download
			_, err = exec.Command("youtube-dl", "-o", "download/"+fileName+".m4a", "-x", "--audio-format", "m4a", ytdl.WebpageURL).Output()
			if err != nil {
				lit.Error("Can't download song: %s", err)
				sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't download song!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, txtChannel)
				return
			}

			// Conversion to DCA
			switch runtime.GOOS {
			case "linux":
				_ = exec.Command("bash", "gen.sh", fileName, fileName+".m4a").Run()
				break
			case "windows":
				_ = exec.Command("gen.bat", fileName, fileName+".m4a").Run()
			}

			err = os.Remove("./download/" + fileName + ".m4a")
			if err != nil {
				lit.Error("Can't delete file: %s", err)
			}
		}

		queue[guildID] = append(queue[guildID], el)
		go playSound(s, guildID, channelID, fileName+".dca", txtChannel)

	}

}

// Searches a song from the query on youtube
func searchDownloadAndPlay(s *discordgo.Session, guildID, channelID, query, user, txtChannel string, random bool) {
	// Gets video id
	out, err := exec.Command("youtube-dl", "--get-id", "ytsearch:\""+query+"\"").Output()
	if err != nil {
		lit.Error("Can't find song on youtube: %s", err)
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No song found!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, txtChannel)
		return
	}

	ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	// Calls download and play for every id we get
	for _, id := range ids {
		downloadAndPlay(s, guildID, channelID, "https://www.youtube.com/watch?v="+id, user, txtChannel, random)
	}

}

// Enqueues song from a spotify playlist, searching them on youtube
func spotifyPlaylist(s *discordgo.Session, guildID, channelID, user, playlistID, txtChannel string, random bool) {

	// We get the playlist from it's link
	playlist, err := client.GetPlaylist(spotify.ID(strings.TrimPrefix(playlistID, "spotify:playlist:")))
	if err != nil {
		lit.Error("Can't get info on a spotify playlist: %s", err)
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about spotify playlist!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, txtChannel)
		return
	}

	// We parse every single song, searching it on youtube
	for _, track := range playlist.Tracks.Tracks {
		go searchDownloadAndPlay(s, guildID, channelID, track.Track.Name+" - "+track.Track.Artists[0].Name, user, txtChannel, random)
	}

}

// Returns song lyrics given a name
func lyrics(song string) []string {
	var lyrics Lyrics

	// Command for downloading lyrics as a JSON file
	cmd := exec.Command("python", "-m", "lyricsgenius", "song", "\""+song+"\"", "--save")

	// We append to the environmental variables the genius token and we run the command
	cmd.Env = append(os.Environ(), "GENIUS_CLIENT_ACCESS_TOKEN="+genius)
	out, err := cmd.Output()
	if err != nil {
		lit.Error("Can't get lyrics for a song: %s", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		// For windows support
		line = strings.TrimSuffix(line, "\r")

		// If the line is the one with the filename
		if strings.HasSuffix(line, ".json.") {
			// We split the line on spaces
			splitted := strings.Split(line, " ")
			// And delete the last dot
			filename := strings.TrimSuffix(splitted[len(splitted)-1], ".")

			// So we open and unmarshal the json file
			file, _ := os.Open(filename)
			byteValue, _ := ioutil.ReadAll(file)
			_ = json.Unmarshal(byteValue, &lyrics)

			// We remove the JSON
			_ = file.Close()
			_ = os.Remove(filename)

			// And return all the lines of the song
			return strings.Split(lyrics.Lyrics, "\n")

		}
	}

	return []string{"No lyrics found"}

}
