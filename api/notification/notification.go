package notification

import (
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

type Notifier struct {
	channels map[snowflake.ID][]chan<- NotificationMessage
	mutex    *sync.RWMutex
}

// NewNotifier creates a new notifier instance
func NewNotifier() *Notifier {
	return &Notifier{
		channels: make(map[snowflake.ID][]chan<- NotificationMessage),
		mutex:    &sync.RWMutex{},
	}
}

// AddChannel adds a channel to the notifier for the given guild
func (n *Notifier) AddChannel(channel chan<- NotificationMessage, guild snowflake.ID) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.channels[guild] = append(n.channels[guild], channel)
}

// RemoveChannel removes a channel from the notifier, closing it in the process and returning true if it was found
func (n *Notifier) RemoveChannel(channel chan<- NotificationMessage, guild snowflake.ID) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for i, c := range n.channels[guild] {
		if c == channel {
			n.channels[guild] = append(n.channels[guild][:i], n.channels[guild][i+1:]...)
			close(c)
			return true
		}
	}

	return false
}

// Notify sends a notification to all channels in the notifier for the given guild (if any)
func (n *Notifier) Notify(guild snowflake.ID, message NotificationMessage) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	for _, c := range n.channels[guild] {
		c <- message
	}
}
