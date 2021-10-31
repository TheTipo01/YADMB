package main

import (
	"io"
	"os/exec"
	"strings"
)

// cmdsStart starts all the exec.Cmd inside the slice
func cmdsStart(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		_ = cmd.Start()
	}
}

// cmdsWait waits for all the exec.Cmd inside the slice to finish processing, to free up resources
func cmdsWait(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		_ = cmd.Wait()
	}
}

// gen substitutes the old scripts, by downloading the song, converting it to DCA and passing it via a pipe
func gen(link string, filename string) (io.ReadCloser, []*exec.Cmd) {
	// Starts yt-dlp with the arguments to select the best audio
	ytDlp := exec.Command("yt-dlp", "-q", "-f", "bestaudio", "-a", "-", "-o", "-")
	ytDlp.Stdin = strings.NewReader(link)
	ytOut, _ := ytDlp.StdoutPipe()

	// We pass it down to ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", "pipe:", "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1")
	ffmpeg.Stdin = ytOut
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	// dca converts it to a format useful for playing back on discord
	dca := exec.Command("dca")
	dca.Stdin = ffmpegOut
	dcaOut, _ := dca.StdoutPipe()

	// tee saves the output from dca to file and also gives it back to us
	tee := exec.Command("tee", "./audio_cache/"+filename+".dca")
	tee.Stdin = dcaOut
	teeOut, _ := tee.StdoutPipe()

	// We give back
	return teeOut, []*exec.Cmd{ytDlp, ffmpeg, dca, tee}
}

// stream substitutes the old scripts for streaming directly to discord from a given source
func stream(link string) (io.ReadCloser, []*exec.Cmd) {
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", link, "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1")
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	dca := exec.Command("dca")
	dca.Stdin = ffmpegOut
	dcaOut, _ := dca.StdoutPipe()

	return dcaOut, []*exec.Cmd{ffmpeg, dca}
}
