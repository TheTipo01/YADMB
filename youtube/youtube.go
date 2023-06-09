package youtube

import (
	"context"
	"github.com/bwmarrin/lit"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"strings"
	"time"
)

type YouTube struct {
	client *youtube.Service
}

func NewYoutube(key string) (*YouTube, error) {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		return nil, err
	}

	return &YouTube{client: youtubeService}, nil
}

// GetVideo returns the title and thumbnail of the given video
func (y *YouTube) GetVideo(id string) *Video {
	response, err := y.client.Videos.List([]string{"snippet", "contentDetails"}).Id(id).Do()
	if err != nil {
		lit.Error("youtube GetVideo: %s", err.Error())
		return nil
	}

	if len(response.Items) == 0 {
		return nil
	}

	duration, _ := time.ParseDuration(strings.TrimPrefix(strings.ToLower(response.Items[0].ContentDetails.Duration), "pt"))

	return &Video{
		Title:     response.Items[0].Snippet.Title,
		Thumbnail: response.Items[0].Snippet.Thumbnails.High.Url,
		URL:       "",
		Duration:  duration.Seconds(),
	}
}

// GetPlaylist returns the URL, title and thumbnail of every element in the given playlist
func (y *YouTube) GetPlaylist(id string) []Video {
	response, err := y.client.PlaylistItems.List([]string{"snippet"}).PlaylistId(id).MaxResults(50).Do()
	if err != nil {
		lit.Error("youtube GetPlaylist: %s", err.Error())
		return nil
	}

	if len(response.Items) == 0 {
		return nil
	}

	result := make([]Video, 0, len(response.Items))
	ids := make([]string, 0, len(response.Items))
	for _, item := range response.Items {
		result = append(result, Video{
			URL:       "https://www.youtube.com/watch?v=" + item.Snippet.ResourceId.VideoId,
			Title:     item.Snippet.Title,
			Thumbnail: item.Snippet.Thumbnails.High.Url,
			Duration:  0,
		})

		ids = append(ids, item.Snippet.ResourceId.VideoId)
	}

	durations := y.getVideosDurations(ids...)
	for i := range result {
		result[i].Duration = durations[i]
	}

	return result
}

func (y *YouTube) getVideosDurations(id ...string) []float64 {
	response, err := y.client.Videos.List([]string{"contentDetails"}).Id(id...).Do()
	if err != nil {
		lit.Error("youtube getVideosDurations: %s", err.Error())
		return nil
	}

	if len(response.Items) == 0 {
		return nil
	}

	result := make([]float64, 0, len(response.Items))
	for _, item := range response.Items {
		duration, _ := time.ParseDuration(strings.TrimPrefix(strings.ToLower(item.ContentDetails.Duration), "pt"))
		result = append(result, duration.Seconds())
	}

	return result
}

// Search returns the URL, title and thumbnail of the first maxResults of the given query
func (y *YouTube) Search(query string, maxResults int64) []Video {
	response, err := y.client.Search.List([]string{"snippet"}).Q(query).Type("video").MaxResults(maxResults).Do()
	if err != nil {
		lit.Error("youtube Search: %s", err.Error())
		return nil
	}

	result := make([]Video, 0, len(response.Items))
	ids := make([]string, 0, len(response.Items))
	for _, item := range response.Items {
		result = append(result, Video{
			URL:       "https://www.youtube.com/watch?v=" + item.Id.VideoId,
			Title:     item.Snippet.Title,
			Thumbnail: item.Snippet.Thumbnails.High.Url,
			Duration:  0,
		})

		ids = append(ids, item.Id.VideoId)
	}

	durations := y.getVideosDurations(ids...)
	for i := range result {
		result[i].Duration = durations[i]
	}

	return result
}
