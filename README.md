# xmasBot
[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/xmasBot)](https://goreportcard.com/report/github.com/TheTipo01/xmasBot)

Discord bot for playing songs h24 in a set of given voice channels.

For downloading songs the bot needs [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [ffmpeg](https://ffmpeg.org/download.html).

There are only two commands, that can be sent privately to the bot (and can only be used by user added to the config file):

- `add <song link>` - Downloads the specified song.
- `restart` - Restarts the bot. Useful if you have just added some songs.
