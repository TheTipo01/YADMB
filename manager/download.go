package manager

import (
	"errors"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/sponsorblock"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/discordgo"
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
func (server *Server) downloadAndPlay(clients *Clients, link, user string, i *discordgo.Interaction, random, loop, respond, priority bool) {
	var c chan struct{}
	if respond {
		c = make(chan struct{})
		go embed.SendEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, link).SetColor(0x7289DA).MessageEmbed, i, c, true)
	}

	link = cleanURL(link)

	// Check if the song is the db, to speedup things
	el, err := clients.Database.CheckInDb(link)
	if err == nil {
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)
		if err == nil && info.Size() > 0 {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.User = user
			el.Reader = f
			el.Closer = f
			el.TextChannel = i.ChannelID
			el.Loop = loop

			if respond {
				go DeleteInteraction(clients.Discord, i, c)
			}
			server.AddSong(priority, el)
			return
		}
	}

	// If we have a valid YouTube client, and the link is a YouTube link, use the YouTube api
	if clients.Youtube != nil && (strings.Contains(link, "youtube.com") || strings.Contains(link, "youtu.be")) {
		err = server.downloadAndPlayYouTubeAPI(clients, link, user, i, random, loop, respond, priority, c)
		// If we have an error, we fall back to yt-dlp
		if err == nil {
			return
		}
	}

	infoJSON, err := getInfo(link)
	if err != nil {
		if respond {
			embed.ModifyInteractionAndDelete(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5)
		} else {
			msg := embed.SendEmbed(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, err.Error()).SetColor(0x7289DA).MessageEmbed, i.ChannelID)
			time.Sleep(time.Second * 5)
			_ = clients.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}

	var ytDLP YtDLP

	// If we want to play the song in a random order, we just shuffle the slice
	if random {
		infoJSON = shuffle(infoJSON)
	}

	if respond {
		go DeleteInteraction(clients.Discord, i, c)
	}

	elements := make([]queue.Element, 0, len(infoJSON))

	// We parse every track as individual json, because yt-dlp
	for _, singleJSON := range infoJSON {
		_ = json.Unmarshal([]byte(singleJSON), &ytDLP)

		el = queue.Element{
			Title:       ytDLP.Title,
			Duration:    FormatDuration(ytDLP.Duration),
			Link:        ytDLP.WebpageURL,
			User:        user,
			Thumbnail:   ytDLP.Thumbnail,
			TextChannel: i.ChannelID,
			Loop:        loop,
		}

		exist := false
		switch ytDLP.Extractor {
		case "youtube":
			el.ID = ytDLP.ID + "-" + ytDLP.Extractor
			// SponsorBlock is supported only on YouTube
			el.Segments = sponsorblock.GetSegments(ytDLP.ID)

			// If the song is on YouTube, we also add it with its compact url, for faster parsing
			clients.Database.AddToDb(el, false)
			exist = true

			el.Link = "https://youtu.be/" + ytDLP.ID
		case "generic":
			// The generic extractor doesn't give out something unique, so we generate one from the link
			el.ID = idGen(el.Link) + "-" + ytDLP.Extractor
		default:
			el.ID = ytDLP.ID + "-" + ytDLP.Extractor
		}

		// We add the song to the db, for faster parsing
		clients.Database.AddToDb(el, exist)

		// If we didn't encounter a playlist, and the link is not the same as the one we got from yt-dlp, add it to the db
		if len(infoJSON) == 1 && el.Link != link {
			go clients.Database.AddLinkDB(el.ID, link)
		}

		// Checks if video is already downloaded
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			pipe, cmd := gen(ytDLP.WebpageURL, el.ID, checkAudioOnly(ytDLP.RequestedFormats))
			el.Reader = pipe
			el.Downloading = true

			el.BeforePlay = func() {
				CmdsStart(cmd)
			}

			el.AfterPlay = func() {
				CmdsWait(cmd)
			}
		} else {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.Reader = f
			el.Closer = f
		}

		elements = append(elements, el)
	}

	server.AddSong(priority, elements...)
}

// DownloadAndPlayYouTubeAPI downloads and plays a song from a YouTube link, parsing the link with the YouTube API
func (server *Server) downloadAndPlayYouTubeAPI(clients *Clients, link, user string, i *discordgo.Interaction, random, loop, respond, priority bool, c chan struct{}) error {
	var (
		result []youtube.Video
		el     queue.Element
	)

	// Check if we have a YouTube playlist, and get its parameter
	u, _ := url.Parse(link)
	q := u.Query()
	if id := q.Get("list"); id != "" {
		result = clients.Youtube.GetPlaylist(id)
	} else {
		if strings.Contains(link, "youtube.com") {
			id = q.Get("v")
		} else {
			id = strings.TrimPrefix(link, "https://youtu.be/")
		}

		if video := clients.Youtube.GetVideo(id); video != nil {
			result = append(result, *video)
		}
	}

	if len(result) == 0 {
		return errors.New("no video found")
	}

	if respond {
		go DeleteInteraction(clients.Discord, i, c)
	}

	elements := make([]queue.Element, 0, len(result))

	for _, r := range result {
		el = queue.Element{
			Title:       r.Title,
			Duration:    FormatDuration(r.Duration),
			Link:        youtubeBase + r.ID,
			User:        user,
			Thumbnail:   r.Thumbnail,
			TextChannel: i.ChannelID,
			Loop:        loop,
		}

		exist := false
		el.ID = r.ID + "-youtube"
		// SponsorBlock is supported only on YouTube
		el.Segments = sponsorblock.GetSegments(r.ID)

		// If the song is on YouTube, we also add it with its compact url, for faster parsing
		clients.Database.AddToDb(el, false)
		exist = true

		// YouTube shorts can have two different links: the one that redirects to a classical YouTube video
		// and one that is played on the new UI. This is a workaround to save also the link to the new UI
		if strings.Contains(link, "shorts") {
			el.Link = link
			clients.Database.AddToDb(el, exist)
		}

		el.Link = "https://youtu.be/" + r.ID

		// We add the song to the db, for faster parsing
		go clients.Database.AddToDb(el, exist)

		// Checks if video is already downloaded
		info, err := os.Stat(constants.CachePath + el.ID + constants.AudioExtension)

		// If not, we download and convert it
		if err != nil || info.Size() <= 0 {
			pipe, cmd := gen(el.Link, el.ID, true)
			el.Reader = pipe
			el.Downloading = true

			el.BeforePlay = func() {
				CmdsStart(cmd)
			}

			el.AfterPlay = func() {
				CmdsWait(cmd)
			}
		} else {
			f, _ := os.Open(constants.CachePath + el.ID + constants.AudioExtension)
			el.Reader = f
			el.Closer = f
		}

		elements = append(elements, el)
	}

	if random {
		rand.Shuffle(len(elements), func(i, j int) {
			elements[i], elements[j] = elements[j], elements[i]
		})
	}

	server.AddSong(priority, elements...)
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
func (server *Server) spotifyPlaylist(clients *Clients, user string, i *discordgo.Interaction, random, loop, priority bool, id spotAPI.ID) {
	if clients.Spotify != nil {
		if playlist, err := clients.Spotify.GetPlaylist(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, "https://open.spotify.com/playlist/"+id.String()).SetColor(0x7289DA).MessageEmbed, i, time.Second*3, true)

			if random {
				rand.Shuffle(len(playlist.Tracks.Tracks), func(i, j int) {
					playlist.Tracks.Tracks[i], playlist.Tracks.Tracks[j] = playlist.Tracks.Tracks[j], playlist.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(playlist.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := playlist.Tracks.Tracks[j]
				link, _ := searchDownloadAndPlay(track.Track.Name+" - "+track.Track.Artists[0].Name, clients.Youtube)
				server.downloadAndPlay(clients, link, user, i, false, loop, false, priority)
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify playlist: %s", err)
			embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
	}
}

func (server *Server) spotifyAlbum(clients *Clients, user string, i *discordgo.Interaction, random, loop, priority bool, id spotAPI.ID) {
	if clients.Spotify != nil {
		if album, err := clients.Spotify.GetAlbum(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.EnqueuedTitle, "https://open.spotify.com/album/"+id.String()).SetColor(0x7289DA).MessageEmbed, i, time.Second*3, true)

			if random {
				rand.Shuffle(len(album.Tracks.Tracks), func(i, j int) {
					album.Tracks.Tracks[i], album.Tracks.Tracks[j] = album.Tracks.Tracks[j], album.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(album.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := album.Tracks.Tracks[j]
				link, _ := searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, clients.Youtube)
				server.downloadAndPlay(clients, link, user, i, false, loop, false, priority)
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify album: %s", err)
			embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
	}
}

// Gets info about a spotify track and plays it, searching it on YouTube
func (server *Server) spotifyTrack(clients *Clients, user string, i *discordgo.Interaction, loop, priority bool, id spotAPI.ID) {
	if clients.Spotify != nil {
		if track, err := clients.Spotify.GetTrack(id); err == nil {
			link, _ := searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, clients.Youtube)
			server.downloadAndPlay(clients, link, user, i, false, loop, true, priority)
		} else {
			lit.Error("Can't get info on a spotify track: %s", err)
			embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error()).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(clients.Discord, embed.NewEmbed().SetTitle(clients.Discord.State.User.Username).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure).SetColor(0x7289DA).MessageEmbed, i, time.Second*5, true)
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
