package main

import (
	"errors"
	"github.com/bwmarrin/lit"
)

const (
	tblSong     = "CREATE TABLE IF NOT EXISTS `song` (`link` varchar(500) NOT NULL, `id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, `thumbnail` varchar(500) NOT NULL, `segments` varchar(1000) NOT NULL, PRIMARY KEY (`link`));"
	tblCommands = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL, `loop` tinyint(1) NOT NULL DEFAULT 0,  PRIMARY KEY (`guild`,`command`));"
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

// Adds a song to the db, so next time we encounter it we don't need to call yt-dlp
func addToDb(el Queue) {
	// We check for empty strings, just to be sure
	if el.link != "" && el.id != "" && el.title != "" && el.duration != "" {
		statement, _ := db.Prepare("INSERT INTO song (link, id, title, duration, thumbnail, segments) VALUES(?, ?, ?, ?, ?, ?)")

		_, err := statement.Exec(el.link, el.id, el.title, el.duration, el.thumbnail, encodeSegments(el.segments))
		if err != nil {
			errStr := err.Error()
			// First error is for SQLite, second one is for MySQL
			if errStr != "constraint failed: UNIQUE constraint failed: song.link (1555)" && errStr != "Error 1062: Duplicate entry '"+el.link+"' for key 'PRIMARY'" {
				lit.Error("Error inserting into the database, %s", err)
			}
		}
	}
}

// Checks if we already have downloaded a song and we've got info about it
func checkInDb(link string) Queue {
	var (
		el              Queue
		encodedSegments string
	)

	el.link = link

	row := db.QueryRow("SELECT * FROM song WHERE link = ?", link)
	_ = row.Scan(&el.link, &el.id, &el.title, &el.duration, &el.thumbnail, &encodedSegments)

	el.segments = decodeSegments(encodedSegments)

	return el
}

// Adds a custom command to db and to the command map
func addCommand(command string, song string, guild string, loop bool) error {
	// If the song is already in the map, we ignore it
	if server[guild].custom[command].link != "" {
		return errors.New("command already exists")
	}

	// Else, we add it to the database
	_, err := db.Exec("INSERT INTO customCommands (guild, command, song, loop) VALUES(?, ?, ?, ?)", guild, command, song, loop)
	if err != nil {
		return errors.New("error inserting into the database: " + err.Error())
	}

	// And the map
	server[guild].custom[command].link = song

	return nil
}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) error {
	// Remove from DB
	if server[guild].custom[command].link == "" {
		return errors.New("command doesn't exist")
	}

	statement, _ := db.Prepare("DELETE FROM customCommands WHERE guild=? AND command=?")
	_, err := statement.Exec(guild, command)
	if err != nil {
		lit.Error("Error removing from the database, %s", err)
	}

	// Remove from the map
	delete(server[guild].custom, command)

	return nil
}

// Loads custom command from the database
func loadCustomCommands() {
	var (
		guild, command, song string
		loop                 bool
	)

	rows, err := db.Query("SELECT * FROM customCommands")
	if err != nil {
		lit.Error("Error querying database, %s", err)
	}

	for rows.Next() {
		err = rows.Scan(&guild, &command, &song, &loop)
		if err != nil {
			lit.Error("Error scanning rows from query, %s", err)
			continue
		}

		// We need to allocate the Server structure before loading custom commands, otherwise we would get a nil pointer deference
		initializeServer(guild)

		server[guild].custom[command] = &CustomCommand{link: song, loop: loop}
	}
}

// Removes an element from the DB
func removeFromDB(el Queue) {
	_, err := db.Exec("DELETE FROM song WHERE id=?", el.id)
	if err != nil {
		lit.Error(err.Error())
	}
}
