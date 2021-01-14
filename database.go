package main

import (
	"errors"
	"github.com/bwmarrin/lit"
	"sync"
)

const (
	tblSong = "CREATE TABLE IF NOT EXISTS `song` (`link` varchar(500) NOT NULL, `id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, `thumbnail` varchar(500) NOT NULL, PRIMARY KEY (`link`));"
	tblCommands = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL,  PRIMARY KEY (`guild`,`command`,`song`));"
)

// Executes a simple query given a DB
func execQuery(query string) {
	statement, err := db.Prepare(query)
	if err != nil {
		lit.Error("Error preparing query, %s", err)
		return
	}

	_, err = statement.Exec()
	if err != nil {
		lit.Error("Error creating table, %s", err)
	}
}

// Adds a song to the db, so next time we encounter it we don't need to call youtube-dl
func addToDb(el Queue) {
	// We check for empty strings, just to be sure
	if el.link != "" && el.id != "" && el.title != "" && el.duration != "" {
		statement, _ := db.Prepare("INSERT INTO song (link, id, title, duration, thumbnail) VALUES(?, ?, ?, ?, ?)")

		_, err := statement.Exec(el.link, el.id, el.title, el.duration, el.thumbnail)
		if err != nil {
			errStr := err.Error()
			// First error is for SQLite, second one is for MySQL
			if errStr != "UNIQUE constraint failed: song.link" && errStr != "Error 1062: Duplicate entry '"+el.link+"' for key 'PRIMARY'" {
				lit.Error("Error inserting into the database, %s", err)
			}
		}
	}
}

// Checks if we already have downloaded a song and we've got info about it
func checkInDb(link string) Queue {
	var el Queue
	el.link = link
	row := db.QueryRow("SELECT * FROM song WHERE link = ?", link)
	_ = row.Scan(&el.link, &el.id, &el.title, &el.duration, &el.thumbnail)

	return el
}

// Adds a custom command to db and to the command map
func addCommand(command string, song string, guild string) error {
	// If the song is already in the map, we ignore it
	if server[guild].custom[command] != "" {
		return errors.New("command already exists")
	}

	// Else, we add it to the map
	server[guild].custom[command] = song

	// And to the database
	statement, _ := db.Prepare("INSERT INTO customCommands (guild, command, song) VALUES(?, ?, ?)")

	_, err := statement.Exec(guild, command, song)
	if err != nil {
		lit.Error("Error inserting into the database, %s", err)
	}

	return nil

}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) {
	// Remove from DB
	statement, _ := db.Prepare("DELETE FROM customCommands WHERE guild=? AND command=?")
	_, err := statement.Exec(guild, command)
	if err != nil {
		lit.Error("Error removing from the database, %s", err)
	}

	// Remove from the map
	delete(server[guild].custom, command)
}

// Loads custom command from the database
func loadCustomCommands() {
	var guild, command, song string

	rows, err := db.Query("SELECT * FROM customCommands")
	if err != nil {
		lit.Error("Error querying database, %s", err)
	}

	for rows.Next() {
		err = rows.Scan(&guild, &command, &song)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		// We need to allocate the Server structure before loading custom commands, otherwise we would get a nil pointer deference
		if server[guild] == nil {
			server[guild] = &Server{server: &sync.Mutex{}, pause: &sync.Mutex{}, custom: make(map[string]string)}
		}

		server[guild].custom[command] = song
	}
}
