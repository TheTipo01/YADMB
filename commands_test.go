package main

import (
	"github.com/bwmarrin/discordgo"
	"testing"
)

// Checks if for every commands there's a function to handle that
func TestCommands(t *testing.T) {
	for _, c := range commands {
		if commandHandlers[c.Name] == nil {
			t.Errorf("Declared command %s in application command slice, but there's no handler.", c.Name)
		}
	}

	for ch := range commandHandlers {
		if !findCommandInCommandHandlers(commands, ch) {
			t.Errorf("Declared command handler %s, but there's no command for it.", ch)
		}
	}
}

func findCommandInCommandHandlers(commands []*discordgo.ApplicationCommand, el string) bool {
	for _, c := range commands {
		if c.Name == el {
			return true
		}
	}

	return false
}
