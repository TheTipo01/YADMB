#!/bin/bash
youtube-dl -o ./download/$2.m4a -f 140 $1
ffmpeg -i ./download/$2.m4a -f s16le -ar 48000 -ac 2 pipe:1 | dca > ./audio_cache/$2.dca