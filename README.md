# YADMB

[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/YADMB)](https://goreportcard.com/report/github.com/TheTipo01/YADMB)

Yet Another Discord Music Bot - A music bot written in go

# Features

- Supports what [yt-dlp](https://github.com/yt-dlp/yt-dlp) supports, plus spotify playlists (if you configure the
  required tokens!)
- Uses slash commands (see [Commands](#commands) for a list of commands)
- Save your favorite songs and playlists with custom commands
- Automatically skips sponsors or moments when there's no music, thanks to
  the [SponsorBlock API](https://sponsor.ajay.app/)
- Caches songs locally, so the bot doesn't have to download them every time
- Stream songs from the internet, useful for radios
- Blacklist users from using the bot
- Allow only certain users to use the bot, with the DJ role

# Requirements

- [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca)
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [ffmpeg](https://ffmpeg.org/download.html)

# Installation

## Natively

See
the [wiki](https://github.com/TheTipo01/YADMB/wiki/Tutorial:-install-YADMB-on-Debian-based-distro-(Raspbian,-Ubuntu...))

## Docker

- Clone the repo
- Modify the `example_config.yml`, by adding your discord bot token (
  see [here](https://github.com/TheTipo01/YADMB/wiki/Creating-and-adding-the-bot-to-your-server) if you don't know how
  to it)
- Rename it in `config.yml` and move it in the `data` directory
- Run `docker-compose up -d`
- Enjoy your YADMB instance!

Note: the docker image is available
on [Docker hub](https://hub.docker.com/r/thetipo01/yadmb), [Quay.io](https://quay.io/repository/thetipo01/yadmb)
and [Github packages](https://github.com/TheTipo01/YADMB/pkgs/container/yadmb).

# Commands

`/play <link> <shuffle> <loop> <priority>` - Plays a song from YouTube or spotify playlist.
If it's not a valid link, it will insert into the queue the first result for the given query.

- `shuffle` if set to true, it will shuffle the playlist.
- `loop` if set to true, it will loop the playlist.
- `priority` if set to true, it will insert the song at the top of the queue.

`/playlist <link> <shuffle> <loop> <priority>` - Same as play, but accepts YouTube playlist

`/skip` - Skips the currently playing song

`/clear` - Clears the entire queue

`/pause` - Pauses currently playing song

`/resume` - Resumes current song

`/queue` - Prints the currently playing song and the next songs

`/disconnect` - Disconnect the bot from the voice channel

`/restart` - Restarts the bot (works only for the bot owner, see `config.yml`)

`/addcustom <custom_command> <song/playlist> <loop>` - Creates a custom command to play a song or playlist. Set `loop`
to true to play the song looped

`/rmcustom <custom_command>` - Removes a custom command from the DB

`/listcustom` - Lists all custom commands for the current server

`/custom <custom_command> <priority>` - Calls a custom command

`/stats` - Stats, like latency, and the size of the local cache

`/goto <time>` - Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...

`/stream <link> <priority>` - Streams a song from the given URL, useful for radios

`/update <link> <info> <song>` - Next time the song `<link>` is played, the bot will (if set to true):

- `<info>` - Update info about the song, like thumbnail, segments from SponsorBlock, title...
- `<song>` - Re-downloads the entire song

`/blacklist <user>` - Adds or remove a person from the bot blacklist

`/dj` - Enables or disable DJ mode for the current server, only users with the DJ role can use the bot

`/djrole <role>` - Sets the DJ role for the current server

`/djroletoggle <user>` - Adds or removes the DJ role from the given user
