package main

import (
	"encoding/json"
	"github.com/bwmarrin/lit"
	"net/http"
	"strconv"
	"strings"
)

// Returns a map for skipping certain frames of a song
func getSegments(videoID string) map[int]bool {

	// Gets segments
	resp, err := http.Get("https://sponsor.ajay.app/api/skipSegments?videoID=" + videoID + "&categories=[\"sponsor\",\"music_offtopic\"]")
	if err != nil {
		lit.Error("Can't get SponsorBlock segments: %s", err)
		return nil
	}

	// If we get the HTTP code 200, segments were found for the given video
	if resp.StatusCode == http.StatusOK {
		var (
			segments   SponsorBlock
			segmentMap = make(map[int]bool)
		)

		err = json.NewDecoder(resp.Body).Decode(&segments)
		_ = resp.Body.Close()
		if err != nil {
			lit.Error("Can't unmarshal JSON, %s", err)
			return nil
		}

		for _, s := range segments {
			if len(s.Segment) == 2 {
				segmentMap[int(s.Segment[0]*frameSeconds)] = true
				segmentMap[int(s.Segment[1]*frameSeconds)] = true
			}
		}

		return segmentMap
	}

	return nil
}

// From a map of segments returns an encoded string
func encodeSegments(segments map[int]bool) string {
	if segments == nil {
		return ""
	}

	var out string

	for k := range segments {
		out += strconv.Itoa(k) + ","
	}

	return strings.TrimSuffix(out, ",")
}

// Decodes segments into a map
func decodeSegments(segments string) map[int]bool {
	if segments == "" {
		return nil
	}

	mapSegments := make(map[int]bool)
	splitted := strings.Split(segments, ",")

	for _, s := range splitted {
		frame, err := strconv.Atoi(s)
		if err == nil {
			mapSegments[frame] = true
		}
	}

	return mapSegments
}
