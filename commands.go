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
					go SendEmbedInteraction(discord.NewEmbed().WithTitle(botName).AddField("Downloading",
						"Please wait...", false).
						WithColor(0x7289DA), e, c, nil)

					err := downloadSong(url)

					<-c
					if err != nil {
						ModifyInteractionAndDelete(discord.NewEmbed().WithTitle(botName).
							AddField("Error", err.Error(), false).
							WithColor(0x7289DA), e, time.Second*10)
					} else {
						ModifyInteractionAndDelete(discord.NewEmbed().WithTitle(botName).
							AddField("Success", "Song added successfully!", false).
							WithColor(0x7289DA), e, time.Second*5)
					}
				} else {
					ModifyInteractionAndDelete(discord.NewEmbed().WithTitle(botName).
						AddField("Error", "Not a valid URL!", false).
						WithColor(0x7289DA), e, time.Second*10)
				}
			} else {
				SendAndDeleteEmbedInteraction(discord.NewEmbed().WithTitle(botName).
					AddField("Error", e.Member().User.Username+" is not in the sudoers file. this incident will be reported", false).
					WithColor(0x7289DA), e, time.Second*10, nil)
			}
		},
	}
)
