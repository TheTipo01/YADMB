package manager

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"

	"layeh.com/gopus"
)

// Opus encoding parameters. They mirror the defaults used by the dca
// executable that this package used to shell out to.
const (
	audioSampleRate = 48000
	audioChannels   = 2
	audioFrameSize  = 960 // 20ms
	audioBitrate    = 64000
)

// pcmFrameBytes is the size in bytes of a single signed 16-bit PCM frame.
const pcmFrameBytes = audioFrameSize * audioChannels * 2

// transcoder is implemented by the audio sources that encode in the background,
// independently of how fast the audio is played back.
type transcoder interface {
	// Start begins the background encoding. It is safe to call it more than once.
	Start()
	// Wait blocks until the background encoding has finished and returns the
	// first error encountered, if any.
	Wait() error
}

// pcmEncoder converts signed 16-bit little-endian PCM into the DCA0 format.
type pcmEncoder struct {
	encoder *gopus.Encoder
	raw     []byte
	samples []int16
	frame   []byte
}

func newPCMEncoder() (*pcmEncoder, error) {
	encoder, err := gopus.NewEncoder(audioSampleRate, audioChannels, gopus.Audio)
	if err != nil {
		return nil, err
	}
	encoder.SetBitrate(audioBitrate)

	return &pcmEncoder{
		encoder: encoder,
		raw:     make([]byte, pcmFrameBytes),
		samples: make([]int16, audioFrameSize*audioChannels),
	}, nil
}

// next reads one PCM frame from r and returns its DCA0 representation. The
// returned slice is only valid until the next call. A short final frame is
// padded with silence; the following call returns io.EOF.
func (e *pcmEncoder) next(r io.Reader) ([]byte, error) {
	n, err := io.ReadFull(r, e.raw)
	if err == io.EOF {
		return nil, io.EOF
	}

	partial := errors.Is(err, io.ErrUnexpectedEOF)
	if err != nil && !partial {
		return nil, err
	}

	// Zero any trailing bytes left over from the previous frame, so that a
	// partial final frame is padded with silence.
	for i := n; i < len(e.raw); i++ {
		e.raw[i] = 0
	}
	for i := range e.samples {
		e.samples[i] = int16(binary.LittleEndian.Uint16(e.raw[i*2:]))
	}

	opus, err := e.encoder.Encode(e.samples, audioFrameSize, pcmFrameBytes)
	if err != nil {
		return nil, err
	}

	e.frame = binary.LittleEndian.AppendUint16(e.frame[:0], uint16(len(opus)))
	e.frame = append(e.frame, opus...)
	return e.frame, nil
}

// dcaReader converts a PCM stream into DCA0 on demand, as it is read. It is used
// for live streams, where buffering the whole input would never terminate.
type dcaReader struct {
	encoder *pcmEncoder
	pcm     io.ReadCloser
	buf     []byte
	offset  int
	err     error
}

// newDCAReader returns a reader that encodes the PCM coming from pcm on demand.
func newDCAReader(pcm io.ReadCloser) *dcaReader {
	return &dcaReader{pcm: pcm}
}

// Read implements io.Reader. It returns the DCA stream, encoding more PCM data
// on demand.
func (r *dcaReader) Read(p []byte) (int, error) {
	for r.offset >= len(r.buf) {
		if r.err != nil {
			return 0, r.err
		}

		if r.encoder == nil {
			encoder, err := newPCMEncoder()
			if err != nil {
				r.finish(err)
				return 0, r.err
			}
			r.encoder = encoder
		}

		frame, err := r.encoder.next(r.pcm)
		if err != nil {
			r.finish(err)
			return 0, r.err
		}

		r.buf = frame
		r.offset = 0
	}

	n := copy(p, r.buf[r.offset:])
	r.offset += n
	return n, nil
}

// Close implements io.Closer. It stops the encoder and releases the source.
func (r *dcaReader) Close() error {
	r.finish(io.EOF)
	return nil
}

func (r *dcaReader) finish(err error) {
	if err == nil {
		err = io.EOF
	}
	if r.err == nil {
		r.err = err
	}
	if r.pcm != nil {
		_ = r.pcm.Close()
		r.pcm = nil
	}
}

// dcaStream decouples encoding from playback. A background goroutine encodes the
// PCM source into the cache file as fast as it can, while the consumer reads
// frames from that same file as they become available. This way a song is
// downloaded and converted at full speed instead of being paced by playback.
type dcaStream struct {
	path string
	pcm  io.ReadCloser

	mu      sync.Mutex
	cond    *sync.Cond
	file    *os.File
	size    int64
	offset  int64
	started bool
	done    bool
	closed  bool
	err     error
}

// newDCAStream returns a stream that encodes pcm into the DCA0 format, saving it
// to path. Nothing happens until Start or Read is called.
func newDCAStream(pcm io.ReadCloser, path string) *dcaStream {
	s := &dcaStream{path: path, pcm: pcm}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// Start begins the background encoding. It is safe to call it more than once.
func (s *dcaStream) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true

	file, err := os.Create(s.path)
	if err != nil {
		s.err = err
		s.done = true
		s.cond.Broadcast()
		s.mu.Unlock()

		_ = s.pcm.Close()
		return
	}
	s.file = file
	s.mu.Unlock()

	go s.encode()
}

// encode is the producer: it reads the PCM source and appends every encoded
// frame to the cache file, without waiting for the consumer.
func (s *dcaStream) encode() {
	var err error

	encoder, err := newPCMEncoder()
	if err == nil {
		var frame []byte
		for {
			frame, err = encoder.next(s.pcm)
			if err != nil {
				break
			}

			if _, err = s.append(frame); err != nil {
				break
			}
		}
	}

	// A regular end of stream is not an error.
	if errors.Is(err, io.EOF) {
		err = nil
	}

	s.mu.Lock()
	if s.closed {
		// The stream was closed on purpose, don't report the resulting error.
		err = nil
	}
	if err != nil && s.err == nil {
		s.err = err
	}
	s.done = true
	s.cond.Broadcast()
	s.mu.Unlock()

	_ = s.pcm.Close()
}

// append writes a frame at the end of the file and wakes up any waiting reader.
func (s *dcaStream) append(frame []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		return 0, io.ErrClosedPipe
	}

	n, err := s.file.WriteAt(frame, s.size)
	s.size += int64(n)
	s.cond.Broadcast()
	return n, err
}

// Read implements io.Reader. It blocks until at least one more frame is
// available, then reads directly from the cache file.
func (s *dcaStream) Read(p []byte) (int, error) {
	s.Start()

	s.mu.Lock()
	for s.offset >= s.size && !s.done {
		s.cond.Wait()
	}
	offset, size, file, err := s.offset, s.size, s.file, s.err
	s.mu.Unlock()

	if offset >= size {
		if err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	if int64(len(p)) > size-offset {
		p = p[:size-offset]
	}

	n, readErr := file.ReadAt(p, offset)
	if n > 0 {
		s.mu.Lock()
		s.offset += int64(n)
		s.mu.Unlock()
	}

	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return n, readErr
	}
	return n, nil
}

// Close implements io.Closer. It stops the background encoding.
func (s *dcaStream) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	return s.pcm.Close()
}

// Wait implements transcoder. It blocks until the background encoding is done
// and closes the cache file.
func (s *dcaStream) Wait() error {
	s.mu.Lock()
	for s.started && !s.done {
		s.cond.Wait()
	}
	err, file := s.err, s.file
	s.mu.Unlock()

	if file != nil {
		_ = file.Close()
	}
	return err
}
