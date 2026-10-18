package embed

import (
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// SendEmbed sends an embed in a given text channel
func SendEmbed(c *bot.Client, embed discord.Embed, txtChannel snowflake.ID) *discord.Message {
	m, err := c.Rest.CreateMessage(txtChannel, discord.NewMessageCreate().AddEmbeds(embed))
	if err != nil || m == nil {
		slog.Error("SendEmbed failed", "error", err)
		return m
	}

	if !TrackMessage(m) {
		_ = c.Rest.DeleteMessage(m.ChannelID, m.ID)
	}

	return m
}

// SendAndDeleteEmbed sends an embed in a given text channel and deletes it after a given duration
func SendAndDeleteEmbed(c *bot.Client, embed discord.Embed, txtChannel snowflake.ID, wait time.Duration) {
	m, err := c.Rest.CreateMessage(txtChannel, discord.NewMessageCreate().AddEmbeds(embed))
	if err != nil || m == nil {
		slog.Error("SendAndDeleteEmbed failed", "error", err)
		return
	}

	if !TrackMessage(m) {
		_ = c.Rest.DeleteMessage(m.ChannelID, m.ID)
		return
	}

	time.Sleep(wait)

	if err := c.Rest.DeleteMessage(m.ChannelID, m.ID); err == nil {
		UntrackMessage(m.ChannelID, m.ID)
	}
}

// SendEmbedInteraction sends an embed as response to an interaction
func SendEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, c chan<- struct{}, isDeferred chan struct{}) {
	if c != nil {
		defer func() { c <- struct{}{} }()
	}

	if e == nil {
		return
	}

	var err error

	if isDeferred != nil {
		<-isDeferred
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdate().AddEmbeds(embed))
	} else {
		err = e.CreateMessage(discord.NewMessageCreate().AddEmbeds(embed))
	}

	if err != nil {
		slog.Error("InteractionRespond failed", "error", err)
		return
	}

	if !TrackInteraction(e.ApplicationID(), e.Token()) {
		_ = e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	}
}

// SendAndDeleteEmbedInteraction sends and deletes after three second an embed in a given channel
func SendAndDeleteEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration, isDeferred chan struct{}) {
	if e == nil {
		return
	}

	SendEmbedInteraction(embed, e, nil, isDeferred)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		slog.Error("InteractionResponseDelete failed", "error", err)
		return
	}

	UntrackInteraction(e.ApplicationID(), e.Token())
}

// Modify an already sent interaction
func ModifyInteraction(e *events.ApplicationCommandInteractionCreate, embed discord.Embed) {
	if e == nil {
		return
	}

	_, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdate().AddEmbeds(embed))
	if err != nil {
		slog.Error("InteractionResponseEdit failed", "error", err)
		return
	}

	if !TrackInteraction(e.ApplicationID(), e.Token()) {
		_ = e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	}
}

// ModifyInteractionAndDelete modifies an already sent interaction and deletes it after the specified wait time
func ModifyInteractionAndDelete(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration) {
	if e == nil {
		return
	}

	ModifyInteraction(e, embed)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		slog.Error("InteractionResponseDelete failed", "error", err)
		return
	}

	UntrackInteraction(e.ApplicationID(), e.Token())
}

func DeferResponse(e *events.ApplicationCommandInteractionCreate) chan struct{} {
	c := make(chan struct{}, 1)

	if e != nil {
		go func() {
			err := e.DeferCreateMessage(false)
			if err == nil {
				if !TrackInteraction(e.ApplicationID(), e.Token()) {
					_ = e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
				}
			}
			c <- struct{}{}
		}()
	} else {
		c <- struct{}{}
	}

	return c
}
