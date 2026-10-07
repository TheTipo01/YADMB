package manager

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"layeh.com/gopus"
)

// dcaFrames splits a DCA0 stream into its individual Opus frames.
func dcaFrames(t *testing.T, data []byte) [][]byte {
	t.Helper()

	var frames [][]byte
	for len(data) > 0 {
		if len(data) < 2 {
			t.Fatalf("truncated DCA length prefix: %d bytes left", len(data))
		}

		n := int(int16(binary.LittleEndian.Uint16(data)))
		data = data[2:]

		if len(data) < n {
			t.Fatalf("truncated DCA frame: want %d bytes, have %d", n, len(data))
		}

		frames = append(frames, data[:n])
		data = data[n:]
	}

	return frames
}

// testPCM generates samples frames of stereo signed 16-bit PCM.
func testPCM(samples int) []byte {
	buf := make([]byte, samples*2*2)
	for i := 0; i < samples; i++ {
		v := int16(math.Sin(2*math.Pi*440*float64(i)/audioSampleRate) * 16000)
		binary.LittleEndian.PutUint16(buf[i*4:], uint16(v))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(v/2))
	}
	return buf
}

func checkFrames(t *testing.T, out []byte, expected int) {
	t.Helper()

	frames := dcaFrames(t, out)
	if len(frames) != expected {
		t.Fatalf("wrong frame count: expected %d, got %d", expected, len(frames))
	}

	dec, err := gopus.NewDecoder(audioSampleRate, audioChannels)
	if err != nil {
		t.Fatalf("NewDecoder failed: %v", err)
	}

	for i, f := range frames {
		if _, err := dec.Decode(f, audioFrameSize, false); err != nil {
			t.Fatalf("frame %d is not valid Opus: %v", i, err)
		}
	}
}

func TestDCAReaderFraming(t *testing.T) {
	// One second of audio is exactly 50 frames of 960 samples.
	r := newDCAReader(io.NopCloser(bytes.NewReader(testPCM(audioSampleRate))))

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	checkFrames(t, out, 50)
}

func TestDCAReaderPartialFrame(t *testing.T) {
	pcm := append(testPCM(audioSampleRate), make([]byte, 100)...)

	r := newDCAReader(io.NopCloser(bytes.NewReader(pcm)))

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	checkFrames(t, out, 51)
}

func TestDCAReaderEmpty(t *testing.T) {
	r := newDCAReader(io.NopCloser(bytes.NewReader(nil)))

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("expected no output, got %d bytes", len(out))
	}
}

func TestDCAStream(t *testing.T) {
	path := filepath.Join(t.TempDir(), "song.dca")
	s := newDCAStream(io.NopCloser(bytes.NewReader(testPCM(audioSampleRate))), path)

	out, err := io.ReadAll(s)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if err := s.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	checkFrames(t, out, 50)

	cached, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cache file was not written: %v", err)
	}
	if !bytes.Equal(out, cached) {
		t.Error("cache file does not match the streamed audio")
	}
}

// TestDCAStreamDownloadsAhead makes sure the whole song is downloaded and
// encoded without the consumer ever reading a single frame.
func TestDCAStreamDownloadsAhead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "song.dca")
	s := newDCAStream(io.NopCloser(bytes.NewReader(testPCM(audioSampleRate))), path)

	s.Start()
	if err := s.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	cached, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cache file was not written: %v", err)
	}

	checkFrames(t, cached, 50)
}

func TestDCAStreamPartialFrame(t *testing.T) {
	path := filepath.Join(t.TempDir(), "song.dca")
	pcm := append(testPCM(audioSampleRate), make([]byte, 100)...)
	s := newDCAStream(io.NopCloser(bytes.NewReader(pcm)), path)

	out, err := io.ReadAll(s)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	checkFrames(t, out, 51)
}
