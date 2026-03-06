package manager

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
)

type dcaFrameProvider struct {
	el        *queue.Element
	server    *Server
	done      chan struct{}
	stop      chan struct{}
	onceClose sync.Once
	skip      bool
}

func newDCAFrameProvider(el *queue.Element, server *Server) *dcaFrameProvider {
	return &dcaFrameProvider{
		el:     el,
		server: server,
		done:   make(chan struct{}),
		stop:   make(chan struct{}),
	}
}

func (d *dcaFrameProvider) ProvideOpusFrame() ([]byte, error) {
	select {
	case <-d.stop:
		return nil, io.EOF
	default:
	}

	select {
	case <-d.server.Pause:
		go notify(notification.NotificationMessage{Notification: notification.Pause, Guild: d.server.GuildID})

		<-d.server.Resume
		go notify(notification.NotificationMessage{Notification: notification.Resume, Guild: d.server.GuildID})
		d.el.Segments = d.server.Queue.GetFirstElement().Segments
	case <-d.stop:
		return nil, io.EOF
	default:
	}

	if _, ok := d.el.Segments[int(d.server.Frames.Load())]; ok {
		d.skip = !d.skip
	}

	var opuslen int16
	err := binary.Read(d.el.Reader, binary.LittleEndian, &opuslen)

	if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
		if d.el.Loop {
			if d.el.Closer != nil {
				_ = d.el.Closer.Close()
			}

			f, err := os.Open(constants.CachePath + d.el.ID + constants.AudioExtension)
			if err != nil {
				d.notifyDone()
				return nil, err
			}
			d.el.Reader = bufio.NewReader(f)
			d.el.Closer = f
			d.server.Frames.Store(0)
			go notify(notification.NotificationMessage{Notification: notification.LoopFinished, Guild: d.server.GuildID})
			return d.ProvideOpusFrame()
		} else {
			d.notifyDone()
			return nil, io.EOF
		}
	}

	if err != nil {
		d.notifyDone()
		_ = os.Remove(constants.CachePath + d.el.ID + constants.AudioExtension)
		return nil, err
	}

	opusData := make([]byte, opuslen)
	err = binary.Read(d.el.Reader, binary.LittleEndian, &opusData)
	if err != nil {
		d.notifyDone()
		_ = os.Remove(constants.CachePath + d.el.ID + constants.AudioExtension)
		return nil, err
	}

	d.server.Frames.Add(1)

	if d.skip {
		return d.ProvideOpusFrame()
	}

	return opusData, nil
}

func (d *dcaFrameProvider) notifyDone() {
	d.onceClose.Do(func() {
		close(d.done)
	})
}

func (d *dcaFrameProvider) Close() {
	close(d.stop)
	d.server.Frames.Store(0)

	if d.el.Closer != nil {
		_ = d.el.Closer.Close()
	}
}
