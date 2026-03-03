package manager

import (
	"bufio"
	"errors"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/embed"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/sponsorblock"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/discord"
	"github.com/goccy/go-json"
	spotAPI "github.com/zmb3/spotify/v2"
)

const youtubeBase = "https://www.youtube.com/watch?v="

// downloadAndPlay downloads and plays a song from a YouTube link
func (server *Server) downloadAndPlay(p PlayEvent, respond bool) {
	var c chan struct{}
	if respond {
		c = make(chan struct{})
		go embed.SendEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, p.Song, false).SetColor(0x7289DA).Build(), p.Event, c, p.IsDeferred)
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
			el.TextChannel = p.Event.Channel().ID()
			el.Loop = p.Loop

			skipTo(p.Song, &el)

			if respond {
				go DeleteInteraction(p.Event.Client(), p.Event, c)
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
			embed.ModifyInteractionAndDelete(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5)
		} else {
			msg := embed.SendEmbed(p.Event.Client(), discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, err.Error(), false).SetColor(0x7289DA).Build(), p.Event.Channel().ID())
			time.Sleep(time.Second * 5)
			_ = p.Event.Client().Rest.DeleteMessage(msg.ChannelID, msg.ID)
		}
		return
	}

	var ytDLP YtDLP

	// If we want to play the song in a random order, we just shuffle the slice
	if p.Random {
		infoJSON = shuffle(infoJSON)
	}

	if respond {
		go DeleteInteraction(p.Event.Client(), p.Event, c)
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
			TextChannel: p.Event.Channel().ID(),
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

			// If the url has a timestamp, we start from that point
			skipTo(p.Song, &el)
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
		// Check if the playlist is in the database
		entries, err := p.Clients.Database.GetPlaylist(id)
		if err == nil && len(entries) > 0 {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, "https://www.youtube.com/playlist?list="+id, false).SetColor(0x7289DA).Build(), p.Event, time.Second*3, p.IsDeferred)

			for j := 0; j < len(entries) && !server.Clear.Load(); j++ {
				p.Song = entries[j]
				server.downloadAndPlay(p, false)
			}

			server.WG.Done()
			return nil
		}

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
		go DeleteInteraction(p.Event.Client(), p.Event, c)
	}

	elements := make([]queue.Element, 0, len(result))

	if len(result) > 1 {
		id := q.Get("list")

		go func() {
			for i := 0; i < len(result); i++ {
				// Add the video as a playlist entry
				err := p.Clients.Database.AddPlaylist(id, youtubeBase+result[i].ID, i)
				if err != nil {
					lit.Error("Error adding playlist to database: %s", err)
				}
			}
		}()
	}

	for _, r := range result {
		el = queue.Element{
			Title:       r.Title,
			Duration:    FormatDuration(r.Duration),
			Link:        youtubeBase + r.ID,
			User:        p.Username,
			Thumbnail:   r.Thumbnail,
			TextChannel: p.Event.Channel().ID(),
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

		// If the url has a timestamp, we start from that point
		checkTimeParameter(q, &el)

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
func searchDownloadAndPlay(query string, yt *youtube.YouTube, db *database.Database) (string, error) {
	// Check if it's in the database
	link, err := db.GetSearch(query)
	if err == nil && link != "" {
		lit.Debug("Found song in database: %s, %s", query, link)
		return link, nil
	}

	if yt != nil {
		result, err := yt.Search(query, 1)
		if err == nil && len(result) > 0 {
			err = db.AddSearch(query, youtubeBase+result[0].ID)
			lit.Debug("Found song from YouTube API, adding to db %s, %s", query, youtubeBase+result[0].ID)
			if err != nil {
				lit.Error("Error adding search to database: %s", err)
			}

			return youtubeBase + result[0].ID, nil
		}
	}

	// yt-dlp is used as a fallback if the YouTube API doesn't return anything or if the YouTube client is not configured
	out, err := exec.Command("yt-dlp", "--get-id", "--quiet", "--ignore-errors", "--no-warnings",
		"--default-search", "ytsearch", query).CombinedOutput()
	if err == nil {
		ids := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")

		if ids[0] != "" {
			err = db.AddSearch(query, youtubeBase+ids[0])
			lit.Debug("Found song from yt-dlp, adding to db %s, %s", query, youtubeBase+ids[0])
			if err != nil {
				lit.Error("Error adding search to database: %s", err)
			}

			return youtubeBase + ids[0], nil
		}
	}

	return "", errors.New("no song found")
}

// Enqueues song from a spotify playlist, searching them on YouTube
func (server *Server) spotifyPlaylist(p PlayEvent, id spotAPI.ID) {
	// Check if we have the playlist in the database
	if entries, err := p.Clients.Database.GetPlaylist(id.String()); err == nil && len(entries) > 0 {
		server.WG.Add(1)

		go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, "https://open.spotify.com/playlist/"+id.String(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*3, p.IsDeferred)

		for j := 0; j < len(entries) && !server.Clear.Load(); j++ {
			p.Song = entries[j]
			server.downloadAndPlay(p, false)
		}

		server.WG.Done()
		return
	}

	if p.Clients.Spotify != nil {
		if playlist, err := p.Clients.Spotify.GetPlaylist(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, "https://open.spotify.com/playlist/"+id.String(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*3, p.IsDeferred)

			if p.Random {
				rand.Shuffle(len(playlist.Tracks.Tracks), func(i, j int) {
					playlist.Tracks.Tracks[i], playlist.Tracks.Tracks[j] = playlist.Tracks.Tracks[j], playlist.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(playlist.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := playlist.Tracks.Tracks[j]

				p.Song, err = searchDownloadAndPlay(track.Track.Name+" - "+track.Track.Artists[0].Name, p.Clients.Youtube, p.Clients.Database)
				if err == nil {
					err = p.Clients.Database.AddPlaylist(id.String(), p.Song, j)
					if err != nil {
						lit.Error("Error adding playlist to database: %s", err)
					}

					server.downloadAndPlay(p, false)
				} else {
					go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
				}
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify playlist: %s", err)
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure, false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
	}
}

func (server *Server) spotifyAlbum(p PlayEvent, id spotAPI.ID) {
	// Check if we have the album in the database
	if entries, err := p.Clients.Database.GetPlaylist(id.String()); err == nil && len(entries) > 0 {
		server.WG.Add(1)

		go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, "https://open.spotify.com/album/"+id.String(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*3, p.IsDeferred)

		for j := 0; j < len(entries) && !server.Clear.Load(); j++ {
			p.Song = entries[j]
			server.downloadAndPlay(p, false)
		}

		server.WG.Done()
		return
	}

	if p.Clients.Spotify != nil {
		if album, err := p.Clients.Spotify.GetAlbum(id); err == nil {
			server.WG.Add(1)

			go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.EnqueuedTitle, "https://open.spotify.com/album/"+id.String(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*3, p.IsDeferred)

			if p.Random {
				rand.Shuffle(len(album.Tracks.Tracks), func(i, j int) {
					album.Tracks.Tracks[i], album.Tracks.Tracks[j] = album.Tracks.Tracks[j], album.Tracks.Tracks[i]
				})
			}

			for j := 0; j < len(album.Tracks.Tracks) && !server.Clear.Load(); j++ {
				track := album.Tracks.Tracks[j]

				p.Song, err = searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, p.Clients.Youtube, p.Clients.Database)
				if err == nil {
					err = p.Clients.Database.AddPlaylist(id.String(), p.Song, j)
					if err != nil {
						lit.Error("Error adding playlist to database: %s", err)
					}

					server.downloadAndPlay(p, false)
				} else {
					go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
				}
			}

			server.WG.Done()
		} else {
			lit.Error("Can't get info on a spotify album: %s", err)
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure, false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
	}
}

// Gets info about a spotify track and plays it, searching it on YouTube
func (server *Server) spotifyTrack(p PlayEvent, id spotAPI.ID) {
	// Check the database for the track
	link, err := p.Clients.Database.GetSearch(id.String())
	if err == nil && link != "" {
		p.Song = link
		server.downloadAndPlay(p, true)

		return
	}

	if p.Clients.Spotify != nil {
		if track, err := p.Clients.Spotify.GetTrack(id); err == nil {
			p.Song, err = searchDownloadAndPlay(track.Name+" - "+track.Artists[0].Name, p.Clients.Youtube, p.Clients.Database)
			if err == nil {
				err = p.Clients.Database.AddSearch(id.String(), p.Song)
				if err != nil {
					lit.Error("Error adding search to database: %s", err)
				}

				server.downloadAndPlay(p, true)
			} else {
				go embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
			}
		} else {
			lit.Error("Can't get info on a spotify track: %s", err)
			embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyError+err.Error(), false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
		}
	} else {
		embed.SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(BotName).AddField(constants.ErrorTitle, constants.SpotifyNotConfigure, false).SetColor(0x7289DA).Build(), p.Event, time.Second*5, p.IsDeferred)
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
