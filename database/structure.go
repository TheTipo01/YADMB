package database

import (
	"sync"

	"github.com/TheTipo01/YADMB/queue"
	"github.com/disgoorg/snowflake/v2"
)

type Database struct {
	AddToDb             func(el queue.Element, exist bool)
	CheckInDb           func(link string) (queue.Element, error)
	AddCommand          func(command string, song string, guild snowflake.ID, loop bool) error
	RemoveCustom        func(command string, guild snowflake.ID) error
	RemoveFromDB        func(el queue.Element)
	GetCustomCommands   func() (map[snowflake.ID]map[string]*CustomCommand, error)
	AddToBlacklist      func(id snowflake.ID) error
	RemoveFromBlacklist func(id snowflake.ID) error
	Close               func()
	UpdateDJRole        func(guild snowflake.ID, role snowflake.ID) error
	GetDJ               func() (map[snowflake.ID]DJ, error)
	GetBlacklist        func() (*sync.Map, error)
	SetDJSettings       func(guild snowflake.ID, enabled bool) error
	AddLinkDB           func(id, link string) error
	GetFavorites        func(userID snowflake.ID) []Favorite
	AddFavorite         func(userID snowflake.ID, favorite Favorite) error
	RemoveFavorite      func(userID snowflake.ID, name string) error
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
	Role    snowflake.ID
}

type Favorite struct {
	Name   string `json:"name"`
	Link   string `json:"link"`
	Folder string `json:"folder"`
}
