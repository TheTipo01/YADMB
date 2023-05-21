package Queue

import (
	"io"
	"sync"
)

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
	// Reader to the song
	Reader io.Reader
	// Closer to the song
	Closer io.Closer
	// Whether we are still downloading the song
	Downloading bool
	// Interaction to edit
	TextChannel string
	// Function to call before playing the song
	BeforePlay func()
	// Function to call after playing the song
	AfterPlay func()
	// Whether to loop the song
	Loop bool
}

type Queue struct {
	queue []Element
	rw    *sync.RWMutex
}

// NewQueue returns a new queue
func NewQueue() Queue {
	return Queue{queue: make([]Element, 0), rw: &sync.RWMutex{}}
}

// IsEmpty returns whether the queue is empty
func (q *Queue) IsEmpty() bool {
	q.rw.RLock()
	defer q.rw.RUnlock()

	return len(q.queue) < 1
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
	defer q.rw.RUnlock()

	queueCopy := make([]Element, len(q.queue))

	for i, el := range q.queue {
		queueCopy[i] = el
	}

	return queueCopy
}

// Clear clears the queue
func (q *Queue) Clear() {
	q.rw.Lock()
	defer q.rw.Unlock()

	q.queue = make([]Element, 0)
}

// ModifyFirstElement modifies the first element in the queue, if it exists
func (q *Queue) ModifyFirstElement(f func(*Element)) {
	q.rw.Lock()
	defer q.rw.Unlock()

	if len(q.queue) > 0 {
		f(&q.queue[0])
	}
}
