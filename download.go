package main

import (
	"encoding/json"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/zmb3/spotify"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Download and plays a song from a youtube link
func downloadAndPlay(s *discordgo.Session, guildID, channelID, link, user string, i *discordgo.Interaction, random bool) {
	c := make(chan int)
	go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(enqueuedTitle, link).SetColor(0x7289DA).MessageEmbed, i, &c)

	// Check if the song is the db, to speedup things
	el := checkInDb(link)
	if el.title != "" {
		info, err := os.Stat(cachePath + el.id + audioExtension)
		if err == nil && info.Size() > 0 {
			el.user = user
			el.channel = channelID
			el.txtChannel = i.ChannelID

			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			go playSound(s, guildID, channelID, el.id+audioExtension, i, nil, &c, nil)
			server[guildID].queueMutex.Unlock()

			return
		}
	}

	splittedOut, err := getInfo(link)
	if err != nil {
		modfyInteractionAndDelete(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
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

		el = Queue{ytdl.Title, formatDuration(ytdl.Duration), "", ytdl.WebpageURL, user, nil, ytdl.Thumbnail, 0, nil, channelID, i.ChannelID}

		exist := false
		switch ytdl.Extractor {
		case "youtube":
			el.id = ytdl.ID + "-" + ytdl.Extractor
			// SponsorBlock is supported only on youtube
			el.segments = getSegments(ytdl.ID)

			// If the song is on YouTube, we also add it with its compact url, for faster parsing
			addToDb(el, false)
			exist = true

			el.link = "https://youtu.be/" + ytdl.ID
		case "generic":
			// The generic extractor doesn't give out something unique, so we generate one from the link
			el.id = idGen(el.link) + "-" + ytdl.Extractor
		default:
			el.id = ytdl.ID + "-" + ytdl.Extractor
		}

		// Checks if video is already downloaded
		info, err := os.Stat(cachePath + el.id + audioExtension)

		// We add the song to the db, for faster parsing
		addToDb(el, exist)

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			pipe, cmd := gen(ytdl.WebpageURL, el.id, checkAudioOnly(ytdl.RequestedFormats))

			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			go playSoundStream(s, guildID, channelID, el.id+audioExtension, i, pipe, cmd)
			server[guildID].queueMutex.Unlock()
		} else {
			server[guildID].queueMutex.Lock()
			server[guildID].queue = append(server[guildID].queue, el)
			go playSound(s, guildID, channelID, el.id+audioExtension, i, nil, &c, nil)
			server[guildID].queueMutex.Unlock()
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
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, nothingFound+splittedOut[len(splittedOut)-1]).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
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
		sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).AddField(errorTitle, spotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
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

			_ = json.NewDecoder(file).Decode(&lyrics)

			// We remove the JSON
			_ = file.Close()
			_ = os.Remove(filename)

			// And return all the lines of the song
			return strings.Split(lyrics.Lyrics, "\n")

		}
	}

	return []string{"No lyrics found"}

}

// getInfo returns info about a song, with every line of the returned array as JSON of type YtDLP
func getInfo(link string) ([]string, error) {
	// Gets info about songs
	out, err := exec.Command("yt-dlp", "--ignore-errors", "-q", "--no-warnings", "-j", link).CombinedOutput()

	// Parse output as string, splitting it on every newline
	splittedOut := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

	if err != nil {
		return nil, errors.New("Can't get info about song: " + splittedOut[len(splittedOut)-1])
	}

	// Check if yt-dlp returned something
	if strings.TrimSpace(splittedOut[0]) == "" {
		return nil, errors.New("yt-dlp returned no songs")
	}

	return splittedOut, nil
}
