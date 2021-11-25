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

// download downloads the song and gives back a pipe with DCA audio
func download(link string) []*exec.Cmd {
	// Starts yt-dlp with the arguments to select the best audio
	ytDlp := exec.Command("yt-dlp", "-q", "-f", "bestaudio*", "-a", "-", "-o", "-", "--geo-bypass")
	ytDlp.Stdin = strings.NewReader(link)
	ytOut, _ := ytDlp.StdoutPipe()

	// We pass it down to ffmpeg
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", "pipe:", "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1", "-af", "loudnorm=I=-16:LRA=11:TP=-1.5")
	ffmpeg.Stdin = ytOut
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	// dca converts it to a format useful for playing back on discord
	dca := exec.Command("dca")
	dca.Stdin = ffmpegOut

	return []*exec.Cmd{ytDlp, ffmpeg, dca}
}

// gen substitutes the old scripts, by downloading the song, converting it to DCA and passing it via a pipe
func gen(link string, filename string) (io.ReadCloser, []*exec.Cmd) {
	cmds := download(link)
	dcaOut, _ := cmds[2].StdoutPipe()

	// tee saves the output from dca to file and also gives it back to us
	tee := exec.Command("tee", cachePath+filename+audioExtension)
	tee.Stdin = dcaOut
	teeOut, _ := tee.StdoutPipe()

	// We give back
	return teeOut, append(cmds, tee)
}

// stream substitutes the old scripts for streaming directly to discord from a given source
func stream(link string) (io.ReadCloser, []*exec.Cmd) {
	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-i", link, "-f", "s16le",
		"-ar", "48000", "-ac", "2", "pipe:1", "-af", "loudnorm=I=-16:LRA=11:TP=-1.5")
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	dca := exec.Command("dca")
	dca.Stdin = ffmpegOut
	dcaOut, _ := dca.StdoutPipe()

	return dcaOut, []*exec.Cmd{ffmpeg, dca}
}
