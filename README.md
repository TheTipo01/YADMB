# YADMB

[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/YADMB)](https://goreportcard.com/report/github.com/TheTipo01/YADMB)

Yet Another Discord Music Bot - A music bot written in go

# Features

- Supports what [yt-dlp](https://github.com/yt-dlp/yt-dlp) does, plus spotify playlists (if you configure the
  required tokens!)
- Uses slash commands (see [Commands](https://thetipo01.github.io/YADMB/commands.html) for a list of commands)
- Save your favorite songs and playlists with custom commands
- Automatically skips sponsors or moments when there's no music, thanks to
  the [SponsorBlock API](https://sponsor.ajay.app/)
- Caches songs locally, so the bot doesn't have to download them every time
- Stream songs from the internet, useful for radios
- Blacklist users from using the bot
- Allow only certain users to use the bot, with the DJ role
- A nice web interface to control the bot

# Requirements

- [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca)
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [ffmpeg](https://ffmpeg.org/download.html)

# Installation

See the [wiki](https://thetipo01.github.io/YADMB/install.html)
