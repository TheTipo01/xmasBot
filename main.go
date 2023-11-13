package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/lit"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/voice/udp"
	"github.com/kkyr/fig"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/voice"
)

var (
	// Discord token
	token string
	// Mutex for downloading songs one at a time
	mutex = &sync.Mutex{}
	// Server map, for holding infos about a server
	servers map[discord.GuildID]*Server
	// Admins holds who are allowed to add songs
	admins map[discord.UserID]bool
	// files holds all the songs
	files []string
	// State
	s       *state.State
	started atomic.Bool
)

const (
	cachePath      = "./audio_cache/"
	audioExtension = ".opus"
)

func init() {
	lit.LogLevel = lit.LogInformational

	var cfg Config
	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		lit.Error(err.Error())
		return
	}

	token = cfg.Token

	servers = make(map[discord.GuildID]*Server, len(cfg.Servers))
	for _, s := range cfg.Servers {
		guild, _ := discord.ParseSnowflake(s.Guild)
		channelSnowflake, _ := discord.ParseSnowflake(s.Channel)

		servers[discord.GuildID(guild)] = &Server{channel: discord.ChannelID(channelSnowflake)}
	}

	admins = make(map[discord.UserID]bool, len(cfg.Admin))
	for _, a := range cfg.Admin {
		snow, err := discord.ParseSnowflake(a)
		userid := discord.UserID(snow)
		if err == nil {
			admins[userid] = true
		}
	}

	// Create folders used by the bot
	if _, err = os.Stat(cachePath); err != nil {
		if err = os.Mkdir(cachePath, 0755); err != nil {
			lit.Error("Cannot create %s, %s", cachePath, err)
		}
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
}

func main() {
	// Create a new Discord session using the provided bot token.
	s = state.New("Bot " + token)
	voice.AddIntents(s)

	h := newHandler(s)
	s.AddInteractionHandler(h)

	if err := cmdroute.OverwriteCommands(h.s, commands); err != nil {
		lit.Error("cannot update commands:", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	s.AddHandler(func(e *gateway.ReadyEvent) {
		if started.CompareAndSwap(false, true) {
			go xmasLoop(s, ctx)
		}

		// Set the status
		_ = s.Gateway().Send(context.Background(), &gateway.UpdatePresenceCommand{
			Activities: []discord.Activity{
				{
					Name: "christmas music",
					Type: discord.ListeningActivity,
				},
			},
			Status: discord.OnlineStatus,
			AFK:    false,
		})
	})

	if err := s.Open(ctx); err != nil {
		log.Fatalln("failed to open:", err)
	}
	defer s.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("xmasBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// Optional constants to tweak the Opus stream.
const (
	frameDuration = 60 // ms
	timeIncrement = 2880
)

func xmasLoop(s *state.State, ctx context.Context) {
	for guild, server := range servers {
		v, err := newVoiceSession(s, ctx, server.channel)
		if err != nil {
			lit.Error("Can't join, %s", err.Error())

			// We can't join the channel, just remove it
			delete(servers, guild)
		} else {
			servers[guild].vs = v
		}
	}

	for {
		for _, v := range rand.Perm(len(files)) {
			playSound(files[v])
		}
	}
}

func newVoiceSession(s *state.State, ctx context.Context, channel discord.ChannelID) (*voice.Session, error) {
	v, err := voice.NewSession(s)
	if err != nil {
		return nil, err
	}

	// Optimize Opus frame duration.
	// This step is optional, but it is recommended.
	v.SetUDPDialer(udp.DialFuncWithFrequency(
		frameDuration*time.Millisecond, // correspond to -frame_duration
		timeIncrement,
	))

	// Join the voice channel.
	err = v.JoinChannelAndSpeak(ctx, channel, false, true)
	return v, err
}

func reconnect(guild discord.GuildID) {
	_ = servers[guild].vs.Leave(context.Background())

	// Recreate the voice session
	servers[guild].vs, _ = newVoiceSession(s, context.Background(), servers[guild].channel)
}
