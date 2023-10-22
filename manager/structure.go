package manager

import (
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/spotify"
	"github.com/TheTipo01/YADMB/youtube"
	"github.com/bwmarrin/discordgo"
	"sync"
	"sync/atomic"
)

// SkipReason is used to determine why playSound returned
type SkipReason int

const (
	Error = iota
	Finished
	Skip
	Clear
)

// Server holds all info about a single server
type Server struct {
	// The queue
	Queue queue.Queue
	// Voice connection
	VC *discordgo.VoiceConnection
	// Custom commands, maps a command to a song
	Custom map[string]*database.CustomCommand
	// Frames
	Frames int
	// Quit channel
	Skip chan SkipReason
	// Whether the job scheduler has started
	Started atomic.Bool
	// Whether to clear the queue
	Clear atomic.Bool
	// Guild ID
	GuildID string
	// Voice channel where the bot is connected
	VoiceChannel string
	// Number of people in the voice channels of the guild
	VoiceChannelMembers map[string]*atomic.Int32
	// Whether the bot is paused
	Paused atomic.Bool
	// Channel for pausing
	Pause chan struct{}
	// Channel for resuming
	Resume chan struct{}
	// Wait group for waiting for spotify to finish before lowering the clear flag
	WG *sync.WaitGroup
	// Role ID for the DJ role
	DjRole string
	// Whether the DJ mode is enabled
	DjMode bool
	// Clients used for interacting with the various APIs
	Clients *Clients
}

// YtDLP structure for holding yt-dlp data
type YtDLP struct {
	Duration         float64          `json:"duration"`
	Thumbnail        string           `json:"thumbnail"`
	Extractor        string           `json:"extractor"`
	ID               string           `json:"id"`
	WebpageURL       string           `json:"webpage_url"`
	Title            string           `json:"title"`
	RequestedFormats RequestedFormats `json:"requested_formats"`
}

// RequestedFormats is used to detect if an audio only codec is available
type RequestedFormats []struct {
	Resolution string `json:"resolution"`
}

// Clients holds all the clients used for interacting with the various APIs
type Clients struct {
	Spotify  *spotify.Spotify
	Youtube  *youtube.YouTube
	Discord  *discordgo.Session
	Database *database.Database
}
