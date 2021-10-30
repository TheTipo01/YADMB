# YADMB
[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/YADMB)](https://goreportcard.com/report/github.com/TheTipo01/YADMB)
[![Build Status](https://app.travis-ci.com/TheTipo01/YADMB.svg?branch=master)](https://app.travis-ci.com/TheTipo01/YADMB)

Yet Another Discord Music Bot - A music bot written in go

# Notes
- We now use slash commands (from release [0.8.0](https://github.com/TheTipo01/YADMB/releases/tag/0.8.0))
- All commands check to see if the caller is in the same voice channel as the bot (only if the skip songs, play a new one)
- Dependencies: [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [yt-dlp](https://github.com/yt-dlp/yt-dlp), [ffmpeg](https://ffmpeg.org/download.html) and [LyricsGenius](https://github.com/johnwmillr/LyricsGenius).
- For tutorials on how to install the bot, see the wiki.
- Uses [SponsorBlock API](https://sponsor.ajay.app/)

## Bot commands

`/play <link>` - Plays a song from YouTube or spotify playlist. If it's not a valid link, it will insert into the queue the first result for the given queue

`/skip` - Skips the currently playing song

`/clear` - Clears the entire queue

`/shuffle <playlist>` - Inserts the song from the given playlist in a random order in the queue

`/pause` - Pauses currently playing song

`/resume` - Resumes current song

`/queue` - Prints the currently playing song and the next songs

`/lyrics <song>` - Tries to search for lyrics of the specified song, or if not specified searches for the title of the currently playing song

`/summon` - Make the bot join your voice channel

`/disconnect` - Disconnect the bot from the voice channel

`/restart` - Restarts the bot (works only for the bot owner, see `config.yml`)

`/addcustom <custom_command> <song/playlist>` - Creates a custom command to play a song or playlist

`/rmcustom <custom_command>` - Removes a custom command from the DB

`/listcustom` - Lists all custom commands for the current server

`/custom <custom_command>` - Calls a custom command

`/stats` - Stats, like latency, and the size of the local cache

`/goto <time>` - Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...

`/stream <link>` - Streams a song from the given URL, useful for radios
