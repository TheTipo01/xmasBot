#!/bin/sh
youtube-dl -f bestaudio -a - -o - | ffmpeg -i pipe: -f s16le -ar 48000 -ac 2 pipe:1 | dca > audio_cache/$1.dca
