package common

import (
	"database/sql"
	"sync"

	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
)

type Common struct {
	db *sql.DB
}

func NewCommon(database *sql.DB) Common {
	return Common{db: database}
}

// CheckInDb checks if we already have downloaded a song, and if we've got info about it
func (c Common) CheckInDb(link string) (queue.Element, error) {
	var (
		el              queue.Element
		encodedSegments string
	)

	err := c.db.QueryRow("SELECT link, songID, title, duration, thumbnail, segments FROM song JOIN link ON songID = id WHERE link = ?", link).
		Scan(&el.Link, &el.ID, &el.Title, &el.Duration, &el.Thumbnail, &encodedSegments)

	if err == nil {
		el.Segments = database.DecodeSegments(encodedSegments)
	}

	return el, err
}

// AddCommand adds a custom command to DB and to the command map
func (c Common) AddCommand(command string, song string, guild string, loop bool) error {
	// Else, we add it to the database
	_, err := c.db.Exec("INSERT INTO customCommands (`guild`, `command`, `song`, `loop`) VALUES(?, ?, ?, ?)", guild, command, song, loop)
	return err
}

// RemoveCustom removes a custom command from the DB and from the command map
func (c Common) RemoveCustom(command string, guild string) error {
	_, err := c.db.Exec("DELETE FROM customCommands WHERE guild=? AND command=?", guild, command)
	if err != nil {
		lit.Error("Error removing from the database, %s", err)
	}

	return nil
}

// GetCustomCommands loads custom command from the database
func (c Common) GetCustomCommands() (map[string]map[string]*database.CustomCommand, error) {
	var (
		command, song, guild string
		loop                 bool
		commands             = make(map[string]map[string]*database.CustomCommand)
	)

	rows, err := c.db.Query("SELECT * FROM customCommands")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&guild, &command, &song, &loop)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		if commands[guild] == nil {
			commands[guild] = make(map[string]*database.CustomCommand)
		}

		commands[guild][command] = &database.CustomCommand{Link: song, Loop: loop}
	}

	return commands, nil
}

// RemoveFromDB removes an element from the DB
func (c Common) RemoveFromDB(el queue.Element) {
	_, err := c.db.Exec("DELETE FROM link WHERE songID=?", el.ID)
	if err != nil {
		lit.Error("Error while deleting from link, %s", err)
	}

	_, err = c.db.Exec("DELETE FROM song WHERE id=?", el.ID)
	if err != nil {
		lit.Error("Error while deleting from song, %s", err)
	}
}

func (c Common) AddToBlacklist(id string) error {
	_, err := c.db.Exec("INSERT INTO blacklist (id) VALUES(?)", id)
	return err
}

func (c Common) RemoveFromBlacklist(id string) error {
	_, err := c.db.Exec("DELETE FROM blacklist WHERE id=?", id)
	return err
}

func (c Common) GetDJ() (map[string]database.DJ, error) {
	roles := make(map[string]database.DJ)

	rows, err := c.db.Query("SELECT guild, role, enabled FROM dj")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var role, guild string
		var enabled bool

		err = rows.Scan(&guild, &role, &enabled)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		roles[guild] = database.DJ{Role: role, Enabled: enabled}
	}

	return roles, nil
}

func (c Common) GetBlacklist() (*sync.Map, error) {
	var ids sync.Map

	rows, err := c.db.Query("SELECT id FROM blacklist")
	if err != nil {
		return &ids, err
	}

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		ids.Store(id, struct{}{})
	}

	return &ids, nil
}

func (c Common) GetFavorites(userID string) []database.Favorite {
	var (
		name, link, folder string
		favorites          = make([]database.Favorite, 0)
	)

	rows, err := c.db.Query("SELECT name, link, folder FROM favorites WHERE userID=?", userID)
	if err != nil {
		lit.Error("Error querying database, %s", err)
		return nil
	}

	for rows.Next() {
		err = rows.Scan(&name, &link, &folder)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		favorites = append(favorites, database.Favorite{Name: name, Link: link, Folder: folder})
	}

	return favorites
}

func (c Common) AddFavorite(userID string, favorite database.Favorite) error {
	_, err := c.db.Exec("INSERT INTO favorites (userID, name, link, folder) VALUES (?, ?, ?, ?)", userID, favorite.Name, favorite.Link, favorite.Folder)
	return err
}

func (c Common) RemoveFavorite(userID, name string) error {
	_, err := c.db.Exec("DELETE FROM favorites WHERE userID=? AND name=?", userID, name)
	return err
}

func (c Common) GetSearch(term string) (string, error) {
	var link string

	err := c.db.QueryRow("SELECT link FROM search WHERE term=?", term).Scan(&link)

	return link, err
}

func (c Common) AddSearch(term, link string) error {
	_, err := c.db.Exec("INSERT INTO search (term, link) VALUES (?, ?)", term, link)
	return err
}

func (c Common) RemoveSearch(term string) error {
	_, err := c.db.Exec("DELETE FROM search WHERE term=?", term)
	return err
}

func (c Common) GetPlaylist(playlist string) ([]string, error) {
	var (
		entry   string
		entries = make([]string, 0)
	)

	rows, err := c.db.Query("SELECT entry FROM playlist WHERE playlist=? ORDER BY number", playlist)
	if err != nil {
		return entries, err
	}

	for rows.Next() {
		err = rows.Scan(&entry)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, err
}

func (c Common) AddPlaylist(playlist, entry string, number int) error {
	_, err := c.db.Exec("INSERT INTO playlist (playlist, entry, number) VALUES (?, ?, ?)", playlist, entry, number)
	if err != nil {
		lit.Error("Error inserting into playlist, %s", err)
		return err
	}

	return nil
}

func (c Common) RemovePlaylist(playlist string) error {
	_, err := c.db.Exec("DELETE FROM playlist WHERE playlist=?", playlist)
	return err
}

func (c Common) Close() {
	_ = c.db.Close()
}
