package main

import (
	"github.com/bwmarrin/discordgo"
	"sync"
)

// Server holds all info about a single server
type Server struct {
	// Mutex for queueing songs correctly
	server *sync.Mutex
	// Mutex for pausing/un-pausing songs
	pause *sync.Mutex
	// Mutex for streaming one song at a time
	stream *sync.Mutex
	// And another one. Used to sync queue read/write
	queueMutex *sync.Mutex
	// Need a boolean to check if song is paused, because the mutex is continuously locked and unlocked
	isPaused bool
	// Variable for skipping a single song
	skip bool
	// Variable for clearing the whole queue
	clear bool
	// The queue
	queue []Queue
	// Voice connection
	vc *discordgo.VoiceConnection
	// Custom commands, maps a command to a song
	custom map[string]*CustomCommand
}

// CustomCommand holds data about a custom command
type CustomCommand struct {
	link string
	loop bool
}

// Queue structure for holding infos about a song
type Queue struct {
	// Title of the song
	title string
	// Duration of the song
	duration string
	// ID of the song
	id string
	// Link of the song
	link string
	// User who requested the song
	user string
	// Message  to delete at the end of the song play
	messageID []discordgo.Message
	// Link to the thumbnail of the video
	thumbnail string
	// We keep how many frame we already played, so we known how many seconds elapsed in the song
	frame int
	// Segments of the song to skip. Uses SponsorBlock API
	segments map[int]bool
	// Channel where we are supposed to play the song. Used for moving the bot around
	channel string
	// Channel for sending error messages and other things
	txtChannel string
}

// YtDLP structure for holding yt-dlp data
type YtDLP struct {
	Duration   float64 `json:"duration"`
	Thumbnail  string  `json:"thumbnail"`
	Extractor  string  `json:"extractor"`
	ID         string  `json:"id"`
	WebpageURL string  `json:"webpage_url"`
	Title      string  `json:"title"`
}

// Lyrics structure for storing lyrics of a song
type Lyrics struct {
	Lyrics string `json:"lyrics"`
}

// SponsorBlock holds data for segments of sponsors in youtube video
type SponsorBlock []struct {
	Category string    `json:"category"`
	Segment  []float64 `json:"segment"`
	UUID     string    `json:"UUID"`
}
