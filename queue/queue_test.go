package queue

import "testing"

func TestQueue_NewQueue(t *testing.T) {
	q := NewQueue()

	if &q == nil {
		t.Error("Expected a new queue, got nil")
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	q := NewQueue()

	if !q.IsEmpty() {
		t.Error("Expected an empty queue, got a non-empty queue")
	}
}

func TestQueue_AddElements(t *testing.T) {
	q := NewQueue()

	q.AddElements(Element{ID: "1"}, Element{ID: "2"}, Element{ID: "3"})

	if q.IsEmpty() {
		t.Error("Expected a non-empty queue, got an empty queue")
	}

	if len(q.Queue) != 3 {
		t.Error("Expected 3, got", len(q.Queue))
	}
}

func TestQueue_GetFirstElement(t *testing.T) {
	q := NewQueue()

	if q.GetFirstElement() != nil {
		t.Error("Expected nil, got an element")
	}

	q.AddElements(Element{ID: "1"}, Element{ID: "2"}, Element{ID: "3"})

	if q.GetFirstElement().ID != "1" {
		t.Error("Expected 1, got", q.GetFirstElement().ID)
	}
}

func TestQueue_AddElementsPriority(t *testing.T) {
	q := NewQueue()

	q.AddElements(Element{ID: "1"}, Element{ID: "2"}, Element{ID: "3"})

	q.AddElementsPriority(Element{ID: "4"})

	if q.GetFirstElement().ID == "4" {
		t.Error("Expected 4, got", q.GetFirstElement().ID)
	}
}

func TestQueue_RemoveFirstElement(t *testing.T) {
	q := NewQueue()

	q.AddElements(Element{ID: "1"}, Element{ID: "2"}, Element{ID: "3"})

	q.RemoveFirstElement()

	if q.GetFirstElement().ID != "2" {
		t.Error("Expected 2, got", q.GetFirstElement().ID)
	}
}
