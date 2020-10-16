# YADMB
[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/YADMB)](https://goreportcard.com/report/github.com/TheTipo01/YADMB)
[![Build Status](https://travis-ci.com/TheTipo01/YADMB.svg?branch=master)](https://travis-ci.com/TheTipo01/YADMB)

Yet Another Discord Music Bot - A music bot written in go

Dependencies: [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [youtube-dl](https://youtube-dl.org/), [ffmpeg](https://ffmpeg.org/download.html) and [LyricsGenius](https://github.com/johnwmillr/LyricsGenius).

For tutorials on how to install the bot, see the wiki.

## Bot commands

`!play <link>` - Plays a song from youtube or spotify playlist

`!shuffle <playlist>` - Shuffles the songs in the playlist and adds them to the queue!pause - Pauses current song

`!resume` - Resumes current song

`!queue` - Returns all the songs in the server queue

`!lyrics <song>` - Tries to search for lyrics of the specified song, or if not specified searches for the title of the currently playing song

`!summon` - Make the bot join your voice channel

`!disconnect` - Disconnect the bot from the voice channel

`!restart` - Restarts the bot

`!custom <custom_command> <song/playlist>` - Creates a custom command to play a song or playlist

`!rmcustom <custom_command>` - Removes a custom command