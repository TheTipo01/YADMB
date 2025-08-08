package main

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

// Checks if for every command there's a function to handle that
func TestCommands(t *testing.T) {
	for _, c := range commands {
		if commandHandlers[c.CommandName()] == nil {
			t.Errorf("Declared command %s in application command slice, but there's no handler.", c.CommandName())
		}
	}

	for ch := range commandHandlers {
		if !findCommandInCommandHandlers(commands, ch) {
			t.Errorf("Declared command handler %s, but there's no command for it.", ch)
		}
	}
}

func findCommandInCommandHandlers(commands []discord.ApplicationCommandCreate, el string) bool {
	for _, c := range commands {
		if c.CommandName() == el {
			return true
		}
	}

	return false
}
