#!/bin/sh
yt-dlp -q -f bestaudio -a - -o - | ffmpeg -hide_banner -loglevel panic -i pipe: -f s16le -ar 48000 -ac 2 pipe:1 | dca | tee audio_cache/$1.dca
