package embed

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"time"
)

// SendEmbed sends an embed in a given text channel
func SendEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) *discordgo.Message {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return nil
	}

	return m
}

// SendEmbedInteraction sends an embed as response to an interaction
func SendEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, c chan<- struct{}, isDeferred bool) {
	// Silently return if the interaction is not valid
	if i.ID == "" {
		return
	}

	var (
		sliceEmbed = []*discordgo.MessageEmbed{embed}
		err        error
	)

	if isDeferred {
		_, err = s.InteractionResponseEdit(i, &discordgo.WebhookEdit{Embeds: &sliceEmbed})
	} else {
		err = s.InteractionRespond(i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Embeds: sliceEmbed}})
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
func SendAndDeleteEmbedInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration, isDeferred bool) {
	// Silently return if the interaction is not valid
	if i.ID == "" {
		return
	}

	SendEmbedInteraction(s, embed, i, nil, isDeferred)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// Modify an already sent interaction
func ModifyInteraction(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction) {
	sliceEmbed := []*discordgo.MessageEmbed{embed}
	_, err := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{Embeds: &sliceEmbed})
	if err != nil {
		lit.Error("InteractionResponseEdit failed: %s", err)
		return
	}
}

// ModifyInteractionAndDelete modifies an already sent interaction and deletes it after the specified wait time
func ModifyInteractionAndDelete(s *discordgo.Session, embed *discordgo.MessageEmbed, i *discordgo.Interaction, wait time.Duration) {
	ModifyInteraction(s, embed, i)

	time.Sleep(wait)

	err := s.InteractionResponseDelete(i)
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}
