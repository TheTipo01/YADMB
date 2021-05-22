#!/bin/sh
ffmpeg -hide_banner -loglevel panic -i $1 -f s16le -ar 48000 -ac 2 pipe:1 | dca
