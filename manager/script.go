package manager

import (
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/TheTipo01/YADMB/constants"
)

// CmdsStart starts all the exec.Cmd inside the slice
func CmdsStart(cmds []*exec.Cmd) []error {
	var errors []error

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// CmdsWait waits for all the exec.Cmd inside the slice to finish processing, to free up resources
func CmdsWait(cmds []*exec.Cmd) []error {
	var errors []error

	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// CmdsKill kills all the exec.Cmd inside the slice
func CmdsKill(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		err := cmd.Process.Kill()
		if err != nil {
			slog.Error("Error killing cmd", "error", err)
		}
	}
}

// download downloads the song and gives back the raw PCM produced by ffmpeg
func download(link string, audioOnly bool) ([]*exec.Cmd, io.ReadCloser) {
	var format string

	// If the flag audioOnly is raised, we use an audio only format to save bandwidth
	if audioOnly {
		format = "bestaudio"
	} else {
		format = "bestaudio*"
	}

	// Starts yt-dlp with the arguments to select the best audio
	ytDlp := exec.Command("yt-dlp", "-q", "-f", format, "-a", "-", "-o", "-", "--geo-bypass")
	ytDlp.Stdin = strings.NewReader(link)
	ytOut, _ := ytDlp.StdoutPipe()

	// We pass it down to ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", "pipe:", "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1", "-af", "loudnorm=I=-16:LRA=11:TP=-1.5")
	ffmpeg.Stdin = ytOut
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	return []*exec.Cmd{ytDlp, ffmpeg}, ffmpegOut
}

// gen substitutes the old scripts, by downloading the song, converting it to DCA and passing it via a pipe
func gen(link string, filename string, audioOnly bool) (io.ReadCloser, []*exec.Cmd) {
	cmds, pcm := download(link, audioOnly)

	// The stream encodes the PCM as fast as it can into the cache file, while
	// playback reads the frames back from there as they become available.
	return newDCAStream(pcm, constants.CachePath+filename+constants.AudioExtension), cmds
}

// Stream substitutes the old scripts for streaming directly to discord from a given source
func Stream(link string) (io.ReadCloser, []*exec.Cmd) {
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", link, "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1", "-af", "loudnorm=I=-16:LRA=11:TP=-1.5")
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	return newDCAReader(ffmpegOut), []*exec.Cmd{ffmpeg}
}
