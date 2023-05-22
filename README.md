# YADMB

[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/YADMB)](https://goreportcard.com/report/github.com/TheTipo01/YADMB)

Yet Another Discord Music Bot - A music bot written in go

# Notes

- We now use slash commands (from release [0.8.0](https://github.com/TheTipo01/YADMB/releases/tag/0.8.0))
- All commands check to see if the caller is in the same voice channel as the bot (only if the skip songs, play a new
  one)
- Dependencies: [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [yt-dlp](https://github.com/yt-dlp/yt-dlp)
  , [ffmpeg](https://ffmpeg.org/download.html).
- For tutorials on how to install the bot, see the wiki.
- Uses [SponsorBlock API](https://sponsor.ajay.app/) to automatically skip sponsors or moments when there's no music
- Normalizes songs

## Bot commands

`/play <link> <playlist> <shuffle> <loop>` - Plays a song from YouTube or spotify playlist. If it's not a valid link, it
will insert into
the queue the first result for the given query.

- `playlist` if set to false/unspecified, it will ignore playlists.
- `shuffle` if set to true, it will shuffle the playlist.
- `loop` if set to true, it will loop the playlist.

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

`/custom <custom_command>` - Calls a custom command

`/stats` - Stats, like latency, and the size of the local cache

`/goto <time>` - Skips to a given time. Valid formats are: 1h10m3s, 3m, 4m10s...

`/stream <link>` - Streams a song from the given URL, useful for radios

`/update <link> <info> <song>` - Next time the song `<link>` is played, the bot will (if set to true):

- `<info>` - Update info about the song, like thumbnail, segments from SponsorBlock, title...
- `<song>` - Redownload the entire song

`/blacklist <user>` - Adds or remove a person from the bot blacklist