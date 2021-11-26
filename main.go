package main

import (
	"fmt"
	"github.com/bwmarrin/lit"
	"github.com/kkyr/fig"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token   string `fig:"token" validate:"required"`
	Guild   string `fig:"guild" validate:"required"`
	Channel string `fig:"channel" validate:"required"`
	Admin   string `fig:"admin" validate:"required"`
}

// Variables used for command line parameters
var (
	token   string
	servers []server
	admin   []string
	mutex   = &sync.Mutex{}
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

	// People that can add songs and restart the bot
	admin = strings.Split(cfg.Admin, ",")

	// Initializing channels to enter
	guilds := strings.Split(cfg.Guild, ",")
	channels := strings.Split(cfg.Channel, ",")

	if len(guilds) != len(channels) {
		lit.Error("Remember to add guilds and channels in pair!")
		return
	}

	for i := range guilds {
		servers = append(servers, server{
			guild:   guilds[i],
			channel: channels[i],
		})
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
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates | discordgo.IntentsDirectMessages)

	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fileInfo, err := ioutil.ReadDir(cachePath)
	if err != nil {
		lit.Error("%s", err)
		return
	}

	go xmasLoop(dg, fileInfo)

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

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || s.State.User.ID == m.Author.ID {
		return
	}

	for _, u := range admin {
		if u == m.Author.ID {
			splitted := strings.Split(m.Content, " ")

			switch splitted[0] {
			case "add":
				err := downloadSong(splitted[1])
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				} else {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Song added successfully!\nRemember to restart the bot to add the song to the queue")
				}
				break
			case "restart":
				_, _ = s.ChannelMessageSend(m.ChannelID, "Restarting...")
				os.Exit(0)
			}
		}
	}

}

func xmasLoop(s *discordgo.Session, fileInfo []os.FileInfo) {
	for i := range servers {
		var err error
		servers[i].vc, err = s.ChannelVoiceJoin(servers[i].guild, servers[i].channel, false, true)
		if err != nil {
			lit.Error("Can't join, %s", err.Error())

			// We can't join the channel, just remove it
			remove(servers, i)
		} else {
			_ = servers[i].vc.Speaking(true)
		}
	}

	for {
		for _, v := range rand.Perm(len(fileInfo)) {
			playSound(fileInfo[v].Name(), s)
		}
	}
}
