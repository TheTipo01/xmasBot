package main

import (
	"fmt"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token   string `fig:"token" validate:"required"`
	Servers []struct {
		Guild   string `fig:"guild" validate:"required"`
		Channel string `fig:"channel" validate:"required"`
	}
	Admin []string `fig:"admin" validate:"required"`
}

type Server struct {
	channel string
	vc      *discordgo.VoiceConnection
}

var (
	// Discord token
	token string
	// Mutex for downloading songs one at a time
	mutex = &sync.Mutex{}
	// Server map, for holding infos about a server
	servers map[string]*Server
	// Admins holds who are allowed to add songs
	admins map[string]bool
	// files holds all the songs
	files []string
)

const (
	cachePath      = "./audio_cache/"
	audioExtension = ".dca"
)

func init() {
	lit.LogLevel = lit.LogInformational

	rand.Seed(time.Now().UnixNano())

	var cfg Config
	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	token = cfg.Token

	servers = make(map[string]*Server, len(cfg.Servers))
	for _, s := range cfg.Servers {
		servers[s.Guild] = &Server{channel: s.Channel}
	}

	admins = make(map[string]bool, len(cfg.Admin))
	for _, a := range cfg.Admin {
		admins[a] = true
	}

	// Create folders used by the bot
	if _, err = os.Stat(cachePath); err != nil {
		if err = os.Mkdir(cachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", cachePath, err)
		}
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// We just need private messages and voiceStates
	dg.Identify.Intents = discordgo.IntentsGuildVoiceStates

	dg.AddHandler(ready)

	// Add commands handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Initial reading
	fileInfo, err := os.ReadDir(cachePath)
	if err != nil {
		lit.Error("%s", err)
		return
	}

	files = make([]string, len(fileInfo))
	for i, f := range fileInfo {
		files[i] = f.Name()
	}

	go xmasLoop(dg)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("xmasBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {
	// Set the playing status.
	err := s.UpdateGameStatus(0, "xmas songs")
	if err != nil {
		lit.Error("Can't set status, %s", err)
	}

	// Checks for unused commands and deletes them
	if cmds, err := s.ApplicationCommands(s.State.User.ID, ""); err == nil {
		found := false

		for _, l := range commands {
			found = false

			for _, o := range cmds {
				// We compare every online command with the ones locally stored, to find if a command with the same name exists
				if l.Name == o.Name {
					// If the options of the command are not equal, we re-register it
					if !isCommandEqual(l, o) {
						lit.Info("Re-registering command `%s`", l.Name)

						_, err = s.ApplicationCommandCreate(s.State.User.ID, "", l)
						if err != nil {
							lit.Error("Cannot create '%s' command: %s", l.Name, err)
						}
					}

					found = true
					break
				}
			}

			// If we didn't found a match for the locally stored command, it means the command is new. We register it
			if !found {
				lit.Info("Registering new command `%s`", l.Name)

				_, err = s.ApplicationCommandCreate(s.State.User.ID, "", l)
				if err != nil {
					lit.Error("Cannot create '%s' command: %s", l.Name, err)
				}
			}
		}
	}
}

func xmasLoop(s *discordgo.Session) {
	for guild, server := range servers {
		var err error
		server.vc, err = s.ChannelVoiceJoin(guild, server.channel, false, true)
		if err != nil {
			lit.Error("Can't join, %s", err.Error())

			// We can't join the channel, just remove it
			delete(servers, guild)
		} else {
			_ = server.vc.Speaking(true)
		}
	}

	for {
		for _, v := range rand.Perm(len(files)) {
			playSound(files[v], s)
		}
	}
}
