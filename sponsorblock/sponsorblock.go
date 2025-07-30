package sponsorblock

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/bwmarrin/lit"
	"github.com/goccy/go-json"
	"net/http"
	"time"
)

const (
	// How many DCA frames are needed for a second. It's not perfect, but good enough.
	frameSeconds = 50.00067787
)

// GetSegments returns a map for skipping certain frames of a song
func GetSegments(videoID string) map[int]struct{} {
	// Gets segments
	req, _ := http.NewRequest("GET", "https://sponsor.ajay.app/api/skipSegments/"+hash(videoID), nil) // Sets timeout to one second, as sometime i
	client := http.Client{Timeout: time.Second}

	req.Header.Set("User-Agent", "github.com/TheTipo01/YADMB")

	q := req.URL.Query()
	q.Set("categories", "[\"sponsor\",\"music_offtopic\"]")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		lit.Error("Can't get SponsorBlock segments: %s", err)
		return nil
	}

	if resp.StatusCode == http.StatusOK {
		var (
			videos     SponsorBlock
			segmentMap map[int]struct{}
		)

		err = json.NewDecoder(resp.Body).Decode(&videos)
		_ = resp.Body.Close()
		if err != nil {
			lit.Error("Can't unmarshal JSON, %s", err)
			return nil
		}

		for _, v := range videos {
			if v.VideoID == videoID {
				segmentMap = make(map[int]struct{}, len(v.Segments)*2)
				for _, segment := range v.Segments {
					if len(segment.Segment) == 2 {
						segmentMap[int(segment.Segment[0]*frameSeconds)] = struct{}{}
						segmentMap[int(segment.Segment[1]*frameSeconds)] = struct{}{}
					}
				}
				return segmentMap
			}
		}
	}

	return nil
}

// returns the first 4 characters of a sha256 hash
func hash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))

	return hex.EncodeToString(h.Sum(nil))[:4]
}
