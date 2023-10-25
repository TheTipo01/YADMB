package notification

import (
	"github.com/TheTipo01/YADMB/queue"
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
)

type Notification int8

type NotificationMessage struct {
	Notification Notification    `json:"notification"`
	Songs        []queue.Element `json:"song,omitempty"`
	Guild        string          `json:"-"`
}