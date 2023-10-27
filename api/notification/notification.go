package notification

import (
	"sync"
)

type Notifier struct {
	channels map[string][]chan<- NotificationMessage
	mutex    *sync.RWMutex
}

func NewNotifier() *Notifier {
	return &Notifier{
		channels: make(map[string][]chan<- NotificationMessage),
		mutex:    &sync.RWMutex{},
	}
}

func (n *Notifier) AddChannel(channel chan<- NotificationMessage, guild string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.channels[guild] = append(n.channels[guild], channel)
}

func (n *Notifier) RemoveChannel(channel chan<- NotificationMessage, guild string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for i, c := range n.channels[guild] {
		if c == channel {
			n.channels[guild] = append(n.channels[guild][:i], n.channels[guild][i+1:]...)
			break
		}
	}
}

func (n *Notifier) Notify(guild string, message NotificationMessage) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	for _, c := range n.channels[guild] {
		c <- message
	}
}
