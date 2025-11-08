package main

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"time"
)

var (
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "add",
			Description: "Adds a song to the bots plethora",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "link",
					Description: "Link of the song to download and play",
					Required:    true,
				},
			},
		},
	}

	// Handler
	commandHandlers = map[string]func(e *events.ApplicationCommandInteractionCreate){
		"add": func(e *events.ApplicationCommandInteractionCreate) {
			if admins[e.Member().User.ID] {
				url := e.SlashCommandInteractionData().String("link")
				if isValidURL(url) {
					c := make(chan struct{})
					go SendEmbedInteraction(discord.NewEmbedBuilder().SetTitle(botName).AddField("Downloading",
						"Please wait...", false).
						SetColor(0x7289DA).Build(), e, c, nil)

					err := downloadSong(url)

					<-c
					if err != nil {
						ModifyInteractionAndDelete(discord.NewEmbedBuilder().SetTitle(botName).
							AddField("Error", err.Error(), false).
							SetColor(0x7289DA).Build(), e, time.Second*10)
					} else {
						ModifyInteractionAndDelete(discord.NewEmbedBuilder().SetTitle(botName).
							AddField("Success", "Song added successfully!", false).
							SetColor(0x7289DA).Build(), e, time.Second*5)
					}
				} else {
					ModifyInteractionAndDelete(discord.NewEmbedBuilder().SetTitle(botName).
						AddField("Error", "Not a valid URL!", false).
						SetColor(0x7289DA).Build(), e, time.Second*10)
				}
			} else {
				SendAndDeleteEmbedInteraction(discord.NewEmbedBuilder().SetTitle(botName).
					AddField("Error", e.Member().User.Username+" is not in the sudoers file. this incident will be reported", false).
					SetColor(0x7289DA).Build(), e, time.Second*10, nil)
			}
		},
	}
)
