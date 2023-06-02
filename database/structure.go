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
}

// CustomCommand holds data about a custom command
type CustomCommand struct {
	Link string
	Loop bool
}
