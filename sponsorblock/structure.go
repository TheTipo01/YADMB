package sponsorblock

// SponsorBlock holds data for segments of sponsors in YouTube video
type SponsorBlock []struct {
	VideoID  string `json:"videoID"`
	Segments []struct {
		Category string    `json:"category"`
		Segment  []float64 `json:"segment"`
	} `json:"segments"`
}
