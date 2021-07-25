package Queue

import "sync"

type Element struct {
	// ID of the song
	ID string
	// Title of the song
	Title string
	// Duration of the song
	Duration string
	// Link of the song
	Link string
	// User who requested the song
	User string
	// Link to the thumbnail of the video
	Thumbnail string
	// Segments of the song to skip. Uses SponsorBlock API
	Segments map[int]bool
}

type Queue struct {
	queue []Element
	rw    *sync.RWMutex
}

// NewQueue returns a new queue
func NewQueue() Queue {
	return Queue{queue: make([]Element, 0), rw: &sync.RWMutex{}}
}

// GetFirstElement returns a copy of the first element in the queue, if it exists
func (q *Queue) GetFirstElement() *Element {
	q.rw.RLock()
	defer q.rw.RUnlock()

	if len(q.queue) < 1 {
		return nil
	}

	top := q.queue[0]
	return &top
}

// AddElements add elements to the queue
func (q *Queue) AddElements(el ...Element) {
	q.rw.Lock()
	defer q.rw.Unlock()

	q.queue = append(q.queue, el...)
}

// RemoveFirstElement removes the first element in the queue, if it exists
func (q *Queue) RemoveFirstElement() {
	q.rw.Lock()
	defer q.rw.Unlock()

	if len(q.queue) > 0 {
		q.queue = q.queue[1:]
	}
}

// GetAllQueue returns a copy of the queue
func (q *Queue) GetAllQueue() []Element {
	q.rw.RLock()
	defer q.rw.Unlock()

	queueCopy := make([]Element, len(q.queue))

	for i, el := range q.queue {
		queueCopy[i] = el
	}

	return queueCopy
}
