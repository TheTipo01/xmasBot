package main

import (
	"context"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

type handler struct {
	*cmdroute.Router
	s *state.State
}

var (
	commands = []api.CreateCommandData{
		{
			Name:        "add",
			Description: "Adds a song to the bots plethora",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "link",
					Description: "Link of the song to download and play",
					Required:    true,
				},
			},
		},
	}
)

func newHandler(s *state.State) *handler {
	h := &handler{s: s}

	h.Router = cmdroute.NewRouter()
	// Automatically defer handles if they're slow.
	h.Use(cmdroute.Deferrable(s, cmdroute.DeferOpts{}))
	h.AddFunc("add", h.cmdAdd)

	return h
}

func (h *handler) cmdAdd(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	var options struct {
		URL string `discord:"link"`
	}
	_ = data.Options.Unmarshal(&options)

	if (data.Event.Member != nil && admins[data.Event.Member.User.ID]) || (data.Event.User != nil && admins[data.Event.User.ID]) {
		if isValidURL(options.URL) {
			err := downloadSong(options.URL)
			if err != nil {
				return &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Error",
							Description: "Error downloading song: " + err.Error(),
							Color:       0x7289DA,
						},
					},
				}
			} else {
				return &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Success",
							Description: "Song added successfully!",
							Color:       0x7289DA,
						},
					},
				}
			}
		} else {
			return &api.InteractionResponseData{
				Embeds: &[]discord.Embed{
					{
						Title:       "Error",
						Description: "Not a valid URL!",
						Color:       0x7289DA,
					},
				},
			}
		}
	} else {
		return &api.InteractionResponseData{
			Embeds: &[]discord.Embed{
				{
					Title:       "Error",
					Description: "You're not in the admin list!",
					Color:       0x7289DA,
				},
			},
		}
	}
}
