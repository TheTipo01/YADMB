#!/bin/bash
ffmpeg -i download/$1.m4a -f s16le -ar 48000 -ac 2 pipe:1 | dca > audio_cache/$1.dca
