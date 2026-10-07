package manager

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDCAStreamWithFFmpeg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not available")
	}

	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:duration=2",
		"-ac", "2", "-f", "s16le", "pipe:1")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()

	path := filepath.Join(t.TempDir(), "song.dca")
	s := newDCAStream(out, path)

	data, err := io.ReadAll(s)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := s.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	frames := dcaFrames(t, data)
	if len(frames) != 100 {
		t.Fatalf("expected 100 frames for 2s, got %d", len(frames))
	}

	cached, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, cached) {
		t.Error("cache file does not match the streamed audio")
	}
}
