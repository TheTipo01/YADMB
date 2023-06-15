package database

import "github.com/TheTipo01/YADMB/queue"

type Database struct {
	AddToDb             func(el queue.Element, exist bool)
	CheckInDb           func(link string) (queue.Element, error)
	AddCommand          func(command string, song string, guild string, loop bool) error
	RemoveCustom        func(command string, guild string) error
	RemoveFromDB        func(el queue.Element)
	GetCustomCommands   func() (map[string]map[string]*CustomCommand, error)
	AddToBlacklist      func(id string) error
	RemoveFromBlacklist func(id string) error
	Close               func()
	UpdateDJRole        func(guild string, role string) error
	GetDJ               func() (map[string]DJ, error)
	GetBlacklist        func() (map[string]bool, error)
	SetDJSettings       func(guild string, enabled bool) error
}

// CustomCommand holds data about a custom command
type CustomCommand struct {
	Link string
	Loop bool
}

type DJ struct {
	Enabled bool
	Role    string
}
