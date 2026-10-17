package notification

import (
	"github.com/TheTipo01/YADMB/queue"
	"github.com/disgoorg/snowflake/v2"
)

const (
	// NewSongs notification for a new song added to the queue
	NewSongs Notification = iota
	// Skip notification for a song being skipped
	Skip
	// Pause notification for a song being paused
	Pause
	// Resume notification for a song being resumed
	Resume
	// Clear notification for the queue being cleared
	Clear
	// Finished notification for the queue being finished
	Finished
	// Playing notification for a song starting to play
	Playing
	// PrioritySong notification for a song being added to the queue with priority
	PrioritySong
	// LoopFinished notification for a song finishing a loop
	LoopFinished
	// Sync notification carrying the current playback position and state
	Sync
)

type Notification int8

type NotificationMessage struct {
	Notification Notification    `json:"notification"`
	Songs        []queue.Element `json:"song,omitempty"`
	// Frames is the number of audio frames already played. Used by Sync.
	Frames int `json:"frames,omitempty"`
	// Playing is set by Sync to tell clients whether playback is currently flowing.
	Playing *bool        `json:"playing,omitempty"`
	Guild   snowflake.ID `json:"-"`
}
