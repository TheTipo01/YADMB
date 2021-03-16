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
	"time"
)

// Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user, txtChannel string, random bool) {

	// Check if the song is the db, to speedup things
	el := checkInDb(link)
	if el.title != "" {
		info, err := os.Stat("./audio_cache/" + el.id + ".dca")
		if err == nil && info.Size() > 0 {
			el.user = user
			el.channel = channelID
			server[guildID].queue = append(server[guildID].queue, el)
			go playSound(s, guildID, channelID, el.id+".dca", txtChannel)
			return
		}
	}

	go sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", link).SetColor(0x7289DA).MessageEmbed, txtChannel, time.Second*5)

	// Gets info about songs
	out, err := exec.Command("youtube-dl", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		lit.Error("Can't get info about song: %s", splittedOut[len(splittedOut)-1])
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about song!\n"+splittedOut[len(splittedOut)-1]).SetColor(0x7289DA).MessageEmbed, txtChannel, time.Second*5)
		return
	}

	// Check if youtube-dl returned something
	if strings.TrimSpace(splittedOut[0]) == "" {
		lit.Error("youtube-dl returned no songs!")
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "youtube-dl returned no songs!").SetColor(0x7289DA).MessageEmbed, txtChannel, time.Second*5)
		return
	}

	var ytdl YoutubeDL

	// If we want to play the song in a random order, we just shuffle the slice
	if random {
		splittedOut = shuffle(splittedOut)
	}

	// We parse every track as individual json, because youtube-dl
	for _, singleJSON := range splittedOut {
		_ = json.Unmarshal([]byte(singleJSON), &ytdl)
		fileName := ytdl.ID + "-" + ytdl.Extractor

		var el Queue
		if ytdl.Extractor == "youtube" {
			el = Queue{ytdl.Title, formatDuration(ytdl.Duration), fileName, ytdl.WebpageURL, user, nil, ytdl.Thumbnail, 0, getSegments(ytdl.ID), channelID}
		} else {
			el = Queue{ytdl.Title, formatDuration(ytdl.Duration), fileName, ytdl.WebpageURL, user, nil, ytdl.Thumbnail, 0, nil, channelID}
		}

		// Checks if video is already downloaded
		info, err := os.Stat("./audio_cache/" + fileName + ".dca")

		// We add the song to the db, for faster parsing
		addToDb(el)

		// If we have a single song, we also add it with the given link
		if len(splittedOut) == 1 {
			el.link = link
			addToDb(el)
		}

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			var cmd *exec.Cmd

			// Download and conversion to DCA
			switch runtime.GOOS {
			case "windows":
				cmd = exec.Command("gen.bat", fileName)
			default:
				cmd = exec.Command("sh", "gen.sh", fileName)
			}

			cmd.Stdin = strings.NewReader(ytdl.WebpageURL)

			pipe, err := cmd.StdoutPipe()
			if err != nil {
				lit.Error("Can't create StdoutPipe: %s", err)
				break
			}

			server[guildID].queue = append(server[guildID].queue, el)
			go playSoundStream(s, guildID, channelID, fileName+".dca", txtChannel, pipe, cmd)
		} else {
			server[guildID].queue = append(server[guildID].queue, el)
			go playSound(s, guildID, channelID, fileName+".dca", txtChannel)
		}

	}
}

// Searches a song from the query on youtube
func searchDownloadAndPlay(s *discordgo.Session, guildID, channelID, query, user, txtChannel string, random bool) {
	// Gets video id
	out, err := exec.Command("youtube-dl", "--get-id", "ytsearch:\""+query+"\"").CombinedOutput()
	if err != nil {
		splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

		lit.Error("Can't find song on youtube: %s", splittedOut[len(splittedOut)-1])
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No song found!\n"+splittedOut[len(splittedOut)-1]).SetColor(0x7289DA).MessageEmbed, txtChannel, time.Second*5)
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
		sendAndDeleteEmbed(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about spotify playlist!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, txtChannel, time.Second*5)
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
