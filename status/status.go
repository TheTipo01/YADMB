package status

import "sync"

type Status struct {
	guildsCount int
	mutex       *sync.RWMutex
}

// NewStatus creates a new Status object
func NewStatus() Status {
	return Status{
		guildsCount: 0,
		mutex:       new(sync.RWMutex),
	}
}

// CompareAndUpdate compares the current guilds count with the given one and updates it if they are different
func (s *Status) CompareAndUpdate(guildsCount int) (bool, int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.guildsCount != guildsCount {
		s.guildsCount = guildsCount
		return true, s.guildsCount
	}
	return false, s.guildsCount
}
