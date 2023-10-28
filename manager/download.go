package manager

import (
	"bufio"
	"errors"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/sponsorblock"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	spotAPI "github.com/zmb3/spotify/v2"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const youtubeBase = "https://www.youtube.com/watch?v="

// downloadAndPlay downloads and plays a song from a YouTube link
func (server *Server) downloadAndPlay(p PlayEvent, respond bool) {
	var c chan struct{}
	if respond {
		c = make(chan struct{})
		go embed.SendEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, p.Song).SetColor(0x7289DA).MessageEmbed, p.Interaction, c, p.IsDeferred)
	}

	p.Song = cleanURL(p.Song)

	// Check if the song is the db, to speedup things
	el, err := p.Clients.Database.CheckInDb(p.Song)
	if err == nil {
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)
		if err == nil && info.Size() > 0 {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.User = p.Username
			el.Reader = bufio.NewReader(f)
			el.Closer = f
			el.TextChannel = p.Interaction.ChannelID
			el.Loop = p.Loop

			if respond {
				go DeleteInteraction(p.Clients.Discord, p.Interaction, c)
			}
			server.AddSong(p.Priority, el)
			return
		}
	}

	// If we have a valid YouTube client, and the link is a YouTube link, use the YouTube api
	if p.Clients.Youtube != nil && (strings.Contains(p.Song, "youtube.com") || strings.Contains(p.Song, "youtu.be")) {
		err = server.downloadAndPlayYouTubeAPI(p, respond, c)
		// If we have an error, we fall back to yt-dlp
		if err == nil {
			return
		}
	}

	infoJSON, err := getInfo(p.Song)
	if err != nil {
		if respond {
			embed.ModifyInteractionAndDelete(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5)
		} else {
			msg := embed.SendEmbed(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction.ChannelID)
			time.Sleep(time.Second * 5)
			_ = p.Clients.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}

	var ytDLP YtDLP

	// If we want to play the song in a random order, we just shuffle the slice
	if p.Random {
		infoJSON = shuffle(infoJSON)
	}

	if respond {
		go DeleteInteraction(p.Clients.Discord, p.Interaction, c)
	}

	elements := make([]queue.Element, 0, len(infoJSON))

	// We parse every track as individual json, because yt-dlp
	for _, singleJSON := range infoJSON {
		_ = json.Unmarshal([]byte(singleJSON), &ytDLP)

		el = queue.Element{
			Title:       ytDLP.Title,
			Duration:    FormatDuration(ytDLP.Duration),
			Link:        ytDLP.WebpageURL,
			User:        p.Username,
			Thumbnail:   ytDLP.Thumbnail,
			TextChannel: p.Interaction.ChannelID,
			Loop:        p.Loop,
		}

		exist := false
		switch ytDLP.Extractor {
		case "youtube":
			el.ID = ytDLP.ID + "-" + ytDLP.Extractor
			// SponsorBlock is supported only on YouTube
			el.Segments = sponsorblock.GetSegments(ytDLP.ID)

			// If the song is on YouTube, we also add it with its compact url, for faster parsing
			p.Clients.Database.AddToDb(el, false)
			exist = true

			el.Link = "https://youtu.be/" + ytDLP.ID
		case "generic":
			// The generic extractor doesn't give out something unique, so we generate one from the link
			el.ID = idGen(el.Link) + "-" + ytDLP.Extractor
		default:
			el.ID = ytDLP.ID + "-" + ytDLP.Extractor
		}

		// We add the song to the db, for faster parsing
		p.Clients.Database.AddToDb(el, exist)

		// If we didn't encounter a playlist, and the link is not the same as the one we got from yt-dlp, add it to the db
		if len(infoJSON) == 1 && el.Link != p.Song {
			go p.Clients.Database.AddLinkDB(el.ID, p.Song)
		}

		// Checks if video is already downloaded
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			pipe, cmd := gen(ytDLP.WebpageURL, el.ID, checkAudioOnly(ytDLP.RequestedFormats))
			el.Reader = bufio.NewReader(pipe)
			el.Downloading = true

			el.BeforePlay = func() {
				CmdsStart(cmd)
			}

			el.AfterPlay = func() {
				CmdsWait(cmd)
			}
		} else {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.Reader = bufio.NewReader(f)
			el.Closer = f
		}

		elements = append(elements, el)
	}

	server.AddSong(p.Priority, elements...)
}

// DownloadAndPlayYouTubeAPI downloads and plays a song from a YouTube link, parsing the link with the YouTube API
func (server *Server) downloadAndPlayYouTubeAPI(p PlayEvent, respond bool, c chan struct{}) error {
	var (
		result []youtube.Video
		el     queue.Element
	)

	// Check if we have a YouTube playlist, and get its parameter
	u, _ := url.Parse(p.Song)
	q := u.Query()
	if id := q.Get("list"); id != "" {
		result = p.Clients.Youtube.GetPlaylist(id)
	} else {
		if strings.Contains(p.Song, "youtube.com") {
			id = q.Get("v")
		} else {
			id = strings.TrimPrefix(p.Song, "https://youtu.be/")
		}

		if video := p.Clients.Youtube.GetVideo(id); video != nil {
			result = append(result, *video)
		}
	}

	if len(result) == 0 {
		return errors.New("no video found")
	}

	if respond {
		go DeleteInteraction(p.Clients.Discord, p.Interaction, c)
	}

	elements := make([]queue.Element, 0, len(result))

	for _, r := range result {
		el = queue.Element{
			Title:       r.Title,
			Duration:    FormatDuration(r.Duration),
			Link:        youtubeBase + r.ID,
			User:        p.Username,
			Thumbnail:   r.Thumbnail,
			TextChannel: p.Interaction.ChannelID,
			Loop:        p.Loop,
		}

		exist := false
		el.ID = r.ID + "-youtube"
		// SponsorBlock is supported only on YouTube
		el.Segments = sponsorblock.GetSegments(r.ID)

		// If the song is on YouTube, we also add it with its compact url, for faster parsing
		p.Clients.Database.AddToDb(el, false)
		exist = true

		// YouTube shorts can have two different links: the one that redirects to a classical YouTube video
		// and one that is played on the new UI. This is a workaround to save also the link to the new UI
		if strings.Contains(p.Song, "shorts") {
			el.Link = p.Song
			p.Clients.Database.AddToDb(el, exist)
		}

		el.Link = "https://youtu.be/" + r.ID

		// We add the song to the db, for faster parsing
		go p.Clients.Database.AddToDb(el, exist)

		// Checks if video is already downloaded
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			pipe, cmd := gen(el.Link, el.ID, true)
			el.Reader = bufio.NewReader(pipe)
			el.Downloading = true

			el.BeforePlay = func() {
				CmdsStart(cmd)
			}

			el.AfterPlay = func() {
				CmdsWait(cmd)
			}
		} else {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.Reader = bufio.NewReader(f)
			el.Closer = f
		}

		elements = append(elements, el)
	}

	if p.Random {
		rand.Shuffle(len(elements), func(i, j int) {
			elements[i], elements[j] = elements[j], elements[i]
		})
	}

	server.AddSong(p.Priority, elements...)
	return nil
}

// Searches a song from the query on YouTube
func searchDownloadAndPlay(query string, yt *youtube.YouTube) (string, error) {
	if yt != nil {
		result := yt.Search(query, 1)
		if len(result) > 0 {
			return youtubeBase + result[0].ID, nil
		}
	} else {
		out, err := exec.Command("yt-dlp", "--get-id", "ytsearch:\""+query+"\"").CombinedOutput()
		if err == nil {
			ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

			if ids[0] != "" {
				return youtubeBase + ids[0], nil
			}
		}
	}

	return "", errors.New("no song found")
}

// Enqueues song from a spotify playlist, searching them on YouTube
func (server *Server) spotifyPlaylist(p PlayEvent, id spotAPI.ID) {
	if p.Clients.Spotify != nil {
		if playlist, err := p.Clients.Spotify.GetPlaylist(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, "https://open.spotify.com/playlist/"+id.String()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*3, p.IsDeferred)

			if p.Random {
				rand.Shuffle(len(playlist.Tracks.Tracks), func(i, j int) {
					playlist.Tracks.Tracks[i], playlist.Tracks.Tracks[j] = playlist.Tracks.Tracks[j], playlist.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(playlist.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := playlist.Tracks.Tracks[j]
				p.Song, _ = searchDownloadAndPlay(track.Track.Name+" - "+track.Track.Artists[0].Name, p.Clients.Youtube)
				server.downloadAndPlay(p, false)
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify playlist: %s", err)
			embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
	}
}

func (server *Server) spotifyAlbum(p PlayEvent, id spotAPI.ID) {
	if p.Clients.Spotify != nil {
		if album, err := p.Clients.Spotify.GetAlbum(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, "https://open.spotify.com/album/"+id.String()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*3, p.IsDeferred)

			if p.Random {
				rand.Shuffle(len(album.Tracks.Tracks), func(i, j int) {
					album.Tracks.Tracks[i], album.Tracks.Tracks[j] = album.Tracks.Tracks[j], album.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(album.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := album.Tracks.Tracks[j]
				p.Song, _ = searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, p.Clients.Youtube)
				server.downloadAndPlay(p, false)
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify album: %s", err)
			embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
	}
}

// Gets info about a spotify track and plays it, searching it on YouTube
func (server *Server) spotifyTrack(p PlayEvent, id spotAPI.ID) {
	if p.Clients.Spotify != nil {
		if track, err := p.Clients.Spotify.GetTrack(id); err == nil {
			p.Song, _ = searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, p.Clients.Youtube)
			server.downloadAndPlay(p, true)
		} else {
			lit.Error("Can't get info on a spotify track: %s", err)
			embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(p.Clients.Discord, embed.NewEmbed().SetTitle(p.Clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, p.Interaction, time.Second*5, p.IsDeferred)
	}
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
