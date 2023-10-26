package queue

import (
	"io"
	"sync"
)

type Element struct {
	// ID of the song
	ID string `json:"id"`
	// Title of the song
	Title string `json:"title"`
	// Duration of the song
	Duration string `json:"duration"`
	// Link of the song
	Link string `json:"link"`
	// User who requested the song
	User string `json:"user"`
	// Link to the thumbnail of the video
	Thumbnail string `json:"thumbnail"`
	// Segments of the song to skip. Uses SponsorBlock API
	Segments map[int]bool `json:"segments,omitempty"`
	// Reader to the song
	Reader io.Reader `json:"-"`
	// Closer to the song
	Closer io.Closer `json:"-"`
	// Whether we are still downloading the song
	Downloading bool `json:"-"`
	// Interaction to edit
	TextChannel string `json:"-"`
	// Function to call before playing the song
	BeforePlay func() `json:"-"`
	// Function to call after playing the song
	AfterPlay func() `json:"-"`
	// Whether to loop the song
	Loop bool `json:"loop"`
	// How many frames have been played. Valid only for the first element in the queue
	Frames int `json:"frames,omitempty"`
	// Whether the song is paused
	IsPaused *bool `json:"isPaused,omitempty"`
}

type Queue struct {
	Queue []Element
	rw    *sync.RWMutex
}

// NewQueue returns a new queue
func NewQueue() Queue {
	return Queue{Queue: make([]Element, 0), rw: &sync.RWMutex{}}
}

// IsEmpty returns whether the queue is empty
func (q *Queue) IsEmpty() bool {
	q.rw.RLock()
	defer q.rw.RUnlock()

	return len(q.Queue) < 1
}

// GetFirstElement returns a copy of the first element in the queue if it exists
func (q *Queue) GetFirstElement() *Element {
	q.rw.RLock()
	defer q.rw.RUnlock()

	if len(q.Queue) < 1 {
		return nil
	}

	top := q.Queue[0]
	return &top
}

// AddElements add elements to the queue
func (q *Queue) AddElements(el ...Element) {
	q.rw.Lock()
	defer q.rw.Unlock()

	q.Queue = append(q.Queue, el...)
}

// AddElementsPriority adds elements to the queue from the second position
// This is useful for adding songs to the top of the queue
// If the queue is empty, it will add the elements to the end of the queue
func (q *Queue) AddElementsPriority(el ...Element) {
	q.rw.Lock()
	defer q.rw.Unlock()

	if len(q.Queue) < 1 {
		q.Queue = append(q.Queue, el...)
	} else {
		q.Queue = append(q.Queue[:1], append(el, q.Queue[1:]...)...)
	}
}

// RemoveFirstElement removes the first element in the queue, if it exists
func (q *Queue) RemoveFirstElement() {
	q.rw.Lock()
	defer q.rw.Unlock()

	if len(q.Queue) > 0 {
		q.Queue = q.Queue[1:]
	}
}

// GetAllQueue returns a copy of the queue
func (q *Queue) GetAllQueue() []Element {
	q.rw.RLock()
	defer q.rw.RUnlock()

	queueCopy := make([]Element, len(q.Queue))

	for i, el := range q.Queue {
		queueCopy[i] = el
	}

	return queueCopy
}

// Clear clears the queue
func (q *Queue) Clear() {
	q.rw.Lock()
	defer q.rw.Unlock()

	q.Queue = make([]Element, 0)
}

// ModifyFirstElement modifies the first element in the queue if it exists
func (q *Queue) ModifyFirstElement(f func(*Element)) {
	q.rw.Lock()
	defer q.rw.Unlock()

	if len(q.Queue) > 0 {
		f(&q.Queue[0])
	}
}
