package main

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/zmb3/spotify"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user string, i *discordgo.Interaction, random bool) {
	c := make(chan int)
	go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Enqueued", link).SetColor(0x7289DA).MessageEmbed, i, &c)

	// Check if the song is the db, to speedup things
	el := checkInDb(link)
	if el.title != "" {
		info, err := os.Stat("./audio_cache/" + el.id + ".dca")
		if err == nil && info.Size() > 0 {
			el.user = user
			el.channel = channelID

			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			server[guildID].queueMutex.Unlock()

			go playSound(s, guildID, channelID, el.id+".dca", i, nil, &c, nil)
			return
		}
	}

	// Gets info about songs
	out, err := exec.Command("yt-dlp", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		lit.Error("Can't get info about song: %s", splittedOut[len(splittedOut)-1])
		modfyInteractionAndDelete(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about song!\n"+splittedOut[len(splittedOut)-1]).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		return
	}

	// Check if yt-dlp returned something
	if strings.TrimSpace(splittedOut[0]) == "" {
		lit.Error("yt-dlp returned no songs!")
		modfyInteractionAndDelete(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "yt-dlp returned no songs!").SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		return
	}

	var ytdl YtDLP

	// If we want to play the song in a random order, we just shuffle the slice
	if random {
		splittedOut = shuffle(splittedOut)
	}

	// We parse every track as individual json, because yt-dlp
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
			pipe, cmd := gen(ytdl.WebpageURL, fileName)

			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			server[guildID].queueMutex.Unlock()

			go playSoundStream(s, guildID, channelID, fileName+".dca", i, pipe, cmd)
		} else {
			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			server[guildID].queueMutex.Unlock()

			go playSound(s, guildID, channelID, fileName+".dca", i, nil, &c, nil)
		}

	}
}

// Searches a song from the query on youtube
func searchDownloadAndPlay(s *discordgo.Session, guildID, channelID, query, user string, i *discordgo.Interaction, random bool) {
	// Gets video id
	out, err := exec.Command("yt-dlp", "--get-id", "ytsearch:\""+query+"\"").CombinedOutput()
	if err != nil {
		splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

		lit.Error("Can't find song on youtube: %s", splittedOut[len(splittedOut)-1])
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "No song found!\n"+splittedOut[len(splittedOut)-1]).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		return
	}

	ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	// Calls download and play for every id we get
	for _, id := range ids {
		downloadAndPlay(s, guildID, channelID, "https://www.youtube.com/watch?v="+id, user, i, random)
	}
}

// Enqueues song from a spotify playlist, searching them on youtube
func spotifyPlaylist(s *discordgo.Session, guildID, channelID, user, playlistID string, i *discordgo.Interaction, random bool) {
	// We get the playlist from it's link
	playlist, err := client.GetPlaylist(spotify.ID(strings.TrimPrefix(playlistID, "spotify:playlist:")))
	if err != nil {
		lit.Error("Can't get info on a spotify playlist: %s", err)
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField("Error", "Can't get info about spotify playlist!\nError code: "+err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		return
	}

	// We parse every single song, searching it on youtube
	for _, track := range playlist.Tracks.Tracks {
		go searchDownloadAndPlay(s, guildID, channelID, track.Track.Name+" - "+track.Track.Artists[0].Name, user, i, random)
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

// gen substitues the old scripts, by downloading the song, converting it to DCA and passing it via a pipe
func gen(link string, filename string) (io.ReadCloser, []*exec.Cmd) {
	// Starts yt-dlp with the arguments to select the best audio
	ytDlp := exec.Command("yt-dlp", "-q", "-f", "bestaudio", "-a", "-", "-o", "-")
	ytDlp.Stdin = strings.NewReader(link)
	ytOut, _ := ytDlp.StdoutPipe()

	// We pass it down to ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", "pipe:", "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1")
	ffmpeg.Stdin = ytOut
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	// dca converts it to a format useful for playing back on discord
	dca := exec.Command("dca")
	dca.Stdin = ffmpegOut
	dcaOut, _ := dca.StdoutPipe()

	// tee saves the output from dca to file and also gives it back to us
	tee := exec.Command("tee", "./audio_cache/"+filename+".dca")
	tee.Stdin = dcaOut
	teeOut, _ := tee.StdoutPipe()

	// We give back
	return teeOut, []*exec.Cmd{ytDlp, ffmpeg, dca, tee}
}
