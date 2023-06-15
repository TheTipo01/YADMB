package common

import (
	"database/sql"
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
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		return err
	}

	return nil
}

func (c Common) RemoveFromBlacklist(id string) error {
	_, err := c.db.Exec("DELETE FROM blacklist WHERE id=?", id)
	if err != nil {
		return err
	}

	return nil
func (c Common) GetBlacklist() (map[string]bool, error) {
	ids := make(map[string]bool)

	rows, err := c.db.Query("SELECT id FROM blacklist")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		ids[id] = true
	}

	return ids, nil
}

func (c Common) Close() {
	_ = c.db.Close()
}
