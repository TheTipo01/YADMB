package main

import (
	"fmt"
	"github.com/TheTipo01/YADMB/Queue"
	"io"
	"os"
	"sync"
)

func NewServer(guildID string) *Server {
	return &Server{
		queue:   Queue.NewQueue(),
		custom:  make(map[string]*CustomCommand),
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
	m.mutex.Lock()
	m.started = true
	m.mutex.Unlock()

	m.clear = false

	for el := m.queue.GetFirstElement(); el != nil && !m.clear; el = m.queue.GetFirstElement() {
		go modifyInteraction(s, NewEmbed().SetTitle(s.State.User.Username).
			AddField("Now playing", fmt.Sprintf("[%s](%s) - %s added by %s", el.Title,
				el.Link, el.Duration, el.User)).
			SetColor(0x7289DA).SetThumbnail(el.Thumbnail).MessageEmbed, m.interaction)

		playSound(m.guildID, el)
		if el.Closer != nil {
			_ = el.Closer.Close()
		}

		// If we are still downloading the song, we need to finish writing it to disk
		if el.Downloading && (!m.clear || m.skip) {
			devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
			_, _ = io.Copy(devnull, el.Reader)
			_ = devnull.Close()
		}

		m.queue.RemoveFirstElement()
	}

	deleteInteraction(s, m.interaction, nil)

	m.mutex.Lock()
	m.started = false
	m.mutex.Unlock()
}

func (m *Server) Clear() {
	m.skip = true
	m.clear = true

	for m.queue.GetFirstElement() != nil {
		m.queue.RemoveFirstElement()
	}
}
