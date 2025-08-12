package embed

import (
	"time"

	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// SendEmbed sends an embed in a given text channel
func SendEmbed(c *bot.Client, embed discord.Embed, txtChannel snowflake.ID) *discord.Message {
	m, _ := c.Rest.CreateMessage(txtChannel, discord.NewMessageCreateBuilder().SetEmbeds(embed).Build())

	return m
}

// SendEmbedInteraction sends an embed as response to an interaction
func SendEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, c chan<- struct{}, isDeferred chan struct{}) {
	var err error

	if isDeferred != nil {
		<-isDeferred
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().SetEmbeds(embed).Build())
	} else {
		err = e.CreateMessage(discord.NewMessageCreateBuilder().SetEmbeds(embed).Build())
	}

	if err != nil {
		lit.Error("InteractionRespond failed: %s", err)
		return
	}

	if c != nil {
		c <- struct{}{}
	}
}

// SendAndDeleteEmbedInteraction sends and deletes after three second an embed in a given channel
func SendAndDeleteEmbedInteraction(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration, isDeferred chan struct{}) {
	SendEmbedInteraction(embed, e, nil, isDeferred)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Modify an already sent interaction
func ModifyInteraction(e *events.ApplicationCommandInteractionCreate, embed discord.Embed) {
	_, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().SetEmbeds(embed).Build())
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}
}

// ModifyInteractionAndDelete modifies an already sent interaction and deletes it after the specified wait time
func ModifyInteractionAndDelete(embed discord.Embed, e *events.ApplicationCommandInteractionCreate, wait time.Duration) {
	ModifyInteraction(e, embed)

	time.Sleep(wait)

	err := e.Client().Rest.DeleteInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

func DeferResponse(e *events.ApplicationCommandInteractionCreate) chan struct{} {
	c := make(chan struct{})
	go func() {
		_ = e.DeferCreateMessage(false)
		c <- struct{}{}
	}()

	return c
}
