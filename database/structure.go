package database

import (
	"sync"

	"github.com/TheTipo01/YADMB/queue"
)

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
	GetBlacklist        func() (*sync.Map, error)
	SetDJSettings       func(guild string, enabled bool) error
	AddLinkDB           func(id, link string) error
	GetFavorites        func(userID string) []Favorite
	AddFavorite         func(userID string, favorite Favorite) error
	RemoveFavorite      func(userID, name string) error
	GetSearch           func(term string) (string, error)
	AddSearch           func(term, link string) error
	RemoveSearch        func(term string) error
	GetPlaylist         func(playlist string) ([]string, error)
	AddPlaylist         func(playlist, entry string, number int) error
	RemovePlaylist      func(playlist string) error
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

type Favorite struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Folder string `json:"folder"`
}
