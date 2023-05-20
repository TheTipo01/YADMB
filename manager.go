package main

import (
	"github.com/TheTipo01/YADMB/Queue"
	"github.com/bwmarrin/lit"
	"sync"
)

func NewServer(guildID string) *Server {
	return &Server{
		queue:   Queue.NewQueue(),
		vc:      nil,
		custom:  make(map[string]*CustomCommand),
		skip:    false,
		clear:   false,
		started: false,
		mutex:   sync.RWMutex{},
		guildID: guildID,
	}
}

func (m *Server) AddSong(el Queue.Element) {
	m.queue.AddElements(el)

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if !m.started {
		go m.play()
	}
}

func (m *Server) play() {
	lit.Debug("Started job for guild %s", m.guildID)
	m.mutex.Lock()
	m.started = true
	m.mutex.Unlock()

	m.clear = false

	for el := m.queue.GetFirstElement(); el != nil && !m.clear; el = m.queue.GetFirstElement() {
		lit.Debug("Playing song: %s", el.ID)
		playSound(m.guildID, el.Reader)
		m.queue.RemoveFirstElement()
	}

	m.mutex.Lock()
	m.started = false
	m.mutex.Unlock()
}

func (m *Server) Clear() {
	m.skip = true
	m.clear = true
}
