package main

import (
	"github.com/TheTipo01/YADMB/manager"
)

func InitializeServer(guild string) {
	if _, ok := server[guild]; !ok {
		server[guild] = manager.NewServer(guild, clients)
	}
}
