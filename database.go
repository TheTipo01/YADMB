package main

import (
	"errors"
	"github.com/TheTipo01/YADMB/Queue"
	"github.com/bwmarrin/lit"
)

// Generic tables, this query can be used on both drivers
const (
	tblSong      = "CREATE TABLE IF NOT EXISTS `song` (`id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, `thumbnail` varchar(500) NOT NULL, `segments` varchar(1000) NOT NULL, PRIMARY KEY (`id`));"
	tblCommands  = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL, `loop` tinyint(1) NOT NULL DEFAULT 0,  PRIMARY KEY (`guild`,`command`));"
	tblBlacklist = "CREATE TABLE IF NOT EXISTS `blacklist`(`id` VARCHAR(18) NOT NULL, PRIMARY KEY (`id`));"
)

// MySQL specific tables
const (
	tblLinkMy = "CREATE TABLE IF NOT EXISTS `link` ( `songID` varchar(200) NOT NULL, `link` varchar(500) NOT NULL, PRIMARY KEY (`link`), KEY `FK__song2` (`songID`), CONSTRAINT `FK__song2` FOREIGN KEY (`songID`) REFERENCES `song` (`id`) );"
)

// SQLite specific tables
const (
	tblLinkLite = "create table if not exists link ( songID varchar(200) not null references song, link varchar(500) not null constraint link_pk primary key );"
)

// Executes a simple query given a DB
func execQuery(query ...string) {
	for _, q := range query {
		_, err := db.Exec(q)
		if err != nil {
			lit.Error("Error executing query, %s", err)
		}
	}
}

// Adds a song to the db, so next time we encounter it we don't need to call yt-dlp
func addToDb(el Queue.Element, exist bool) {
	// We check for empty strings, just to be sure
	if el.Link != "" && el.ID != "" && el.Title != "" && el.Duration != "" {
		if !exist {
			_, err := db.Exec("INSERT "+ignoreType+" IGNORE INTO song (id, title, duration, thumbnail, segments) VALUES (?, ?, ?, ?, ?)",
				el.ID, el.Title, el.Duration, el.Thumbnail, encodeSegments(el.Segments))
			if err != nil {
				lit.Error("Error inserting into song, %s", err)
			}
		}

		_, err := db.Exec("INSERT "+ignoreType+" IGNORE INTO link (songID, link) VALUES (?, ?)", el.ID, el.Link)
		if err != nil {
			lit.Error("Error inserting into link, %s", err)
		}
	}
}

// Checks if we already have downloaded a song, and if we've got info about it
func checkInDb(link string) (Queue.Element, error) {
	var (
		el              Queue.Element
		encodedSegments string
	)

	err := db.QueryRow("SELECT link, songID, title, duration, thumbnail, segments FROM song JOIN link ON songID = id WHERE link = ?", link).
		Scan(&el.Link, &el.ID, &el.Title, &el.Duration, &el.Thumbnail, &encodedSegments)

	if err == nil {
		el.Segments = decodeSegments(encodedSegments)
	}

	return el, err
}

// Adds a custom command to db and to the command map
func addCommand(command string, song string, guild string, loop bool) error {
	// If the song is already in the map, we ignore it
	if server[guild].custom[command] != nil {
		return errors.New("command already exists")
	}

	// Else, we add it to the database
	_, err := db.Exec("INSERT INTO customCommands (`guild`, `command`, `song`, `loop`) VALUES(?, ?, ?, ?)", guild, command, song, loop)
	if err != nil {
		return errors.New("error inserting into the database: " + err.Error())
	}

	// And the map
	server[guild].custom[command] = &CustomCommand{link: song, loop: loop}

	return nil
}

// Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) error {
	// Remove from DB
	if server[guild].custom[command] == nil {
		return errors.New("command doesn't exist")
	}

	_, err := db.Exec("DELETE FROM customCommands WHERE guild=? AND command=?", guild, command)
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
func removeFromDB(el Queue.Element) {
	_, err := db.Exec("DELETE FROM link WHERE songID=?", el.ID)
	if err != nil {
		lit.Error("Error while deleting from link, %s", err)
	}

	_, err = db.Exec("DELETE FROM song WHERE id=?", el.ID)
	if err != nil {
		lit.Error("Error while deleting from song, %s", err)
	}
}
