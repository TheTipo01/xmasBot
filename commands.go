package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "add",
			Description: "Adds a song to the bots plethora",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link of the song to download and play",
					Required:    true,
				},
			},
		},
	}

	// Handler
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if (i.User != nil && admins[i.User.ID]) || (i.Member != nil && admins[i.Member.User.ID]) {
				url := i.ApplicationCommandData().Options[0].StringValue()
				if isValidURL(url) {
					c := make(chan int)
					go sendEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).
						AddField("Downloading", "Please wait...").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, &c)

					err := downloadSong(url)

					<-c
					if err != nil {
						modifyInteractionAndDelete(s, NewEmbed().SetTitle(s.State.User.Username).
							AddField("Error", err.Error()).
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10)
					} else {
						modifyInteractionAndDelete(s, NewEmbed().SetTitle(s.State.User.Username).
							AddField("Success", "Song added successfully!").
							SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*5)
					}
				} else {
					sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).
						AddField("Error", "Not a valid URL!").
						SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10)
				}
			} else {
				var user string
				if i.User != nil {
					user = i.User.Username
				} else {
					user = i.Member.User.Username
				}
				
				sendAndDeleteEmbedInteraction(s, NewEmbed().SetTitle(s.State.User.Username).
					AddField("Error", user+" is not in the sudoers file. this incident will be reported").
					SetColor(0x7289DA).MessageEmbed, i.Interaction, time.Second*10)
			}
		},
	}
)
