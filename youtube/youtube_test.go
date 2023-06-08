package youtube

import (
	"os"
	"strings"
	"testing"
)

func TestYouTube_GetVideo(t *testing.T) {
	yt, err := NewYoutube(os.Getenv("YOUTUBE_API"))
	if err != nil {
		t.Fatal(err)
	}

	response := yt.GetVideo("dQw4w9WgXcQ")
	if response == nil {
		t.Fatal("video response is nil")
	}

	if !strings.Contains(strings.ToLower(response.Title), "never gonna give you up") {
		t.Fatal("video title does not match")
	}

	result := yt.GetPlaylist("PLqcP8b8B-T6D9cru6eqsWQRLc8LlkIAKh")
	if result == nil {
		t.Fatal("playlist response is nil")
	}

	if result[0].URL != "https://www.youtube.com/watch?v=PvuYSybooLg" {
		t.Fatal("first video does not match")
	}
}
