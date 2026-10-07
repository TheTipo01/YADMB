package manager

import (
	"encoding/binary"
	"errors"
	"io"
	"os"

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

// dcaReader converts a stream of signed 16-bit little-endian PCM audio into the
// DCA0 format used for Discord voice playback. It replaces the external dca
// executable, encoding the audio on demand as it is read.
type dcaReader struct {
	// encoder is created lazily on the first frame, so that no Opus state is
	// allocated for audio that is never played.
	encoder *gopus.Encoder

	// pcm is the source of raw PCM audio, usually ffmpeg's stdout.
	pcm io.Reader
	// source is closed once the stream ends, so that the process feeding it can
	// terminate instead of blocking on a full pipe.
	source io.Closer

	// cache, when not empty, is the path the encoded audio is saved to while it
	// is being streamed. The file is created lazily on the first frame.
	cache     string
	cacheFile *os.File

	// raw and samples are scratch buffers reused for every encoded frame.
	raw     []byte
	samples []int16

	// buf holds the DCA frame currently being handed out, made of a 2 byte
	// little-endian length followed by the encoded Opus data.
	buf    []byte
	offset int

	// pending marks that a final, partial frame has been produced and that an
	// EOF should be returned once it has been consumed.
	pending bool
	err     error
}

// newDCAReader returns a reader that encodes the PCM coming from pcm into the
// DCA0 format. When cache is not empty, the encoded audio is also written to
// that path.
func newDCAReader(pcm io.ReadCloser, cache string) *dcaReader {
	return &dcaReader{
		pcm:     pcm,
		source:  pcm,
		cache:   cache,
		raw:     make([]byte, pcmFrameBytes),
		samples: make([]int16, audioFrameSize*audioChannels),
	}
}

// Read implements io.Reader. It returns the DCA stream, encoding more PCM data
// on demand.
func (r *dcaReader) Read(p []byte) (int, error) {
	for r.offset >= len(r.buf) {
		if r.err != nil {
			return 0, r.err
		}

		if r.pending {
			r.finish(io.EOF)
			return 0, r.err
		}

		if err := r.readFrame(); err != nil {
			r.finish(err)
			return 0, r.err
		}
	}

	n := copy(p, r.buf[r.offset:])
	r.offset += n
	return n, nil
}

// Close implements io.Closer. It stops the encoder and releases the underlying
// source and cache file.
func (r *dcaReader) Close() error {
	r.pending = false
	r.finish(io.EOF)
	return nil
}

// readFrame reads one PCM frame, encodes it and stores the resulting DCA frame
// in buf. It returns io.EOF when the source has no more data.
func (r *dcaReader) readFrame() error {
	n, err := io.ReadFull(r.pcm, r.raw)
	if err == io.EOF {
		return io.EOF
	}

	partial := errors.Is(err, io.ErrUnexpectedEOF)
	if err != nil && !partial {
		return err
	}

	// Zero any trailing bytes left over from the previous frame, so that a
	// partial final frame is padded with silence.
	for i := n; i < len(r.raw); i++ {
		r.raw[i] = 0
	}
	for i := range r.samples {
		r.samples[i] = int16(binary.LittleEndian.Uint16(r.raw[i*2:]))
	}

	if r.encoder == nil {
		r.encoder, err = gopus.NewEncoder(audioSampleRate, audioChannels, gopus.Audio)
		if err != nil {
			return err
		}
		r.encoder.SetBitrate(audioBitrate)
	}

	opus, err := r.encoder.Encode(r.samples, audioFrameSize, pcmFrameBytes)
	if err != nil {
		return err
	}

	r.buf = binary.LittleEndian.AppendUint16(r.buf[:0], uint16(len(opus)))
	r.buf = append(r.buf, opus...)
	r.offset = 0

	if err := r.writeCache(); err != nil {
		return err
	}

	// If this was a partial frame, signal the end of the stream once it has been
	// consumed.
	r.pending = partial
	return nil
}

// writeCache lazily opens the cache file and appends the current frame to it.
func (r *dcaReader) writeCache() error {
	if r.cache == "" {
		return nil
	}

	if r.cacheFile == nil {
		f, err := os.Create(r.cache)
		if err != nil {
			return err
		}
		r.cacheFile = f
	}

	_, err := r.cacheFile.Write(r.buf)
	return err
}

// finish records the first error encountered, closes the source and flushes the
// cache file. A nil error is treated as a regular end of stream.
func (r *dcaReader) finish(err error) {
	if err == nil {
		err = io.EOF
	}

	if r.err == nil {
		r.err = err
	}

	if r.source != nil {
		_ = r.source.Close()
		r.source = nil
	}

	if r.cacheFile != nil {
		if cacheErr := r.cacheFile.Close(); cacheErr != nil && r.err == io.EOF {
			r.err = cacheErr
		}
		r.cacheFile = nil
	}
}
