package youtube

import (
	"os"
	"strings"
	"testing"
)

var yt *YouTube

func init() {
	var err error

	yt, err = NewYoutube(os.Getenv("YOUTUBE_API"))
	if err != nil {
		panic(err)
	}
}

func TestYouTube_GetVideo(t *testing.T) {
	response := yt.GetVideo("dQw4w9WgXcQ")
	if response == nil {
		t.Fatal("video response is nil")
	}

	if !strings.Contains(strings.ToLower(response.Title), "never gonna give you up") {
		t.Fatal("video title does not match")
	}
}

func TestYouTube_GetPlaylist(t *testing.T) {
	result := yt.GetPlaylist("PLqcP8b8B-T6D9cru6eqsWQRLc8LlkIAKh")
	if result == nil {
		t.Fatal("playlist response is nil")
	}

	if result[0].URL != "https://www.youtube.com/watch?v=PvuYSybooLg" {
		t.Fatal("first video does not match")
	}
}

func TestYouTube_Search(t *testing.T) {
	result := yt.Search("never gonna give you up", 1)
	if result == nil {
		t.Fatal("search response is nil")
	}

	if result[0].URL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Fatal("first video does not match")
	}
}
