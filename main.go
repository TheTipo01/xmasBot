package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/kkyr/fig"
)

var (
	// Discord token
	token string
	// Mutex for downloading songs one at a time
	mutex = &sync.Mutex{}
	// Server map, for holding infos about a server
	servers map[snowflake.ID]*Server
	// Admins holds who are allowed to add songs
	admins map[snowflake.ID]bool
	// files holds all the songs
	files []string
	// Bot status
	status string
	// Bon name
	botName string
)

const (
	cachePath      = "./audio_cache/"
	audioExtension = ".dca"
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
	status = cfg.Status

	servers = make(map[snowflake.ID]*Server, len(cfg.Servers))
	for _, s := range cfg.Servers {
		servers[snowflake.MustParse(s.Channel)] = &Server{channel: snowflake.MustParse(s.Channel)}
	}

	admins = make(map[snowflake.ID]bool, len(cfg.Admin))
	for _, a := range cfg.Admin {
		admins[snowflake.MustParse(a)] = true
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

	files = make([]string, 0, len(fileInfo))
	for _, f := range fileInfo {
		if name := f.Name(); strings.HasSuffix(name, ".dca") {
			files = append(files, name)
		}
	}
}

func main() {
	client, _ := disgo.New(token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildVoiceStates)),

		bot.WithEventListenerFunc(ready),
		bot.WithEventListenerFunc(interactionCreate),

		bot.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))),
	)

	if err := client.OpenGateway(context.TODO()); err != nil {
		lit.Error("errors while connecting to gateway %s", err)
		return
	}

	// Register commands
	_, err := client.Rest.SetGlobalCommands(client.ApplicationID, commands)
	if err != nil {
		lit.Error("Error registering commands: %s", err)
		return
	}

	go xmasLoop(client)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("xmasBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	client.Close(context.TODO())
}

func ready(e *events.Ready) {
	client := e.Client()
	err := client.SetPresence(context.TODO(), gateway.WithListeningActivity(status))
	if err != nil {
		lit.Error("Error setting status: %s", err)
	}

	botName = e.User.Username
}

func xmasLoop(client *bot.Client) {
	for guild, server := range servers {
		server.vc = client.VoiceManager.CreateConn(guild)

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		if err := server.vc.Open(ctx, server.channel, false, false); err != nil {
			lit.Error("Can't join, %s", err.Error())

			// We can't join the channel, just remove it
			delete(servers, guild)
			continue
		}

		if err := server.vc.SetSpeaking(ctx, voice.SpeakingFlagMicrophone); err != nil {
			lit.Error("error setting speaking flag: %s", err.Error())
		}
	}

	for {
		for _, v := range rand.Perm(len(files)) {
			playSound(files[v])
		}
	}
}

func interactionCreate(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	// Ignores commands from DM
	if e.Context() == discord.InteractionContextTypeGuild {
		if h, ok := commandHandlers[data.CommandName()]; ok {
			h(e)
		}
	} else {
		SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(botName).AddField("Error",
			"Don't use the bot in private!", false).
			SetColor(0x7289DA).Build(), e, time.Second*15, nil)
	}
}
