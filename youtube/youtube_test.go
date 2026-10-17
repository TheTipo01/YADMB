package youtube

import (
	"os"
	"strings"
	"testing"
)

var yt *YouTube

func init() {
	key := os.Getenv("YOUTUBE_API")
	if key == "" {
		return
	}

	var err error

	yt, err = NewYoutube(key)
	if err != nil {
		panic(err)
	}
}

func requireYouTube(t *testing.T) {
	t.Helper()

	if yt == nil {
		t.Skip("YOUTUBE_API is not set")
	}
}

func TestYouTube_GetVideo(t *testing.T) {
	requireYouTube(t)

	response := yt.GetVideo("dQw4w9WgXcQ")
	if response == nil {
		t.Fatal("video response is nil")
	}

	if !strings.Contains(strings.ToLower(response.Title), "never gonna give you up") {
		t.Fatal("video title does not match")
	}
}

func TestYouTube_GetPlaylist(t *testing.T) {
	requireYouTube(t)

	result := yt.GetPlaylist("PLqcP8b8B-T6D9cru6eqsWQRLc8LlkIAKh")
	if result == nil {
		t.Fatal("playlist response is nil")
	}

	if result[0].ID != "PvuYSybooLg" {
		t.Fatal("first video does not match")
	}
}

func TestYouTube_Search(t *testing.T) {
	requireYouTube(t)

	result, err := yt.Search("never gonna give you up", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("search response is nil")
	}

	if result[0].ID != "dQw4w9WgXcQ" {
		t.Fatal("first video does not match")
	}
}
