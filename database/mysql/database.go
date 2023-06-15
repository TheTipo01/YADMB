package mysql

import (
	"database/sql"
	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/database/common"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
	_ "github.com/go-sql-driver/mysql"
)

const (
	tblSong      = "CREATE TABLE IF NOT EXISTS `song` (`id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, `thumbnail` varchar(500) NOT NULL, `segments` varchar(1000) NOT NULL, PRIMARY KEY (`id`));"
	tblCommands  = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL, `loop` tinyint(1) NOT NULL DEFAULT 0,  PRIMARY KEY (`guild`,`command`));"
	tblBlacklist = "CREATE TABLE IF NOT EXISTS `blacklist`(`id` VARCHAR(20) NOT NULL, PRIMARY KEY (`id`));"
	tblLink      = "CREATE TABLE IF NOT EXISTS `link` ( `songID` varchar(200) NOT NULL, `link` varchar(500) NOT NULL, PRIMARY KEY (`link`), KEY `FK__song2` (`songID`), CONSTRAINT `FK__song2` FOREIGN KEY (`songID`) REFERENCES `song` (`id`) );"
)

var db *sql.DB

// NewDatabase returns a new database
func NewDatabase(dsn string) *database.Database {
	var err error

	// Open database connection
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		lit.Error("Error opening db connection, %s", err)
		return nil
	}

	// Create tables if they don't exist
	database.ExecQuery(db, tblSong, tblCommands, tblBlacklist, tblLink)

	c := common.NewCommon(db)

	return &database.Database{
		AddToDb:             addToDb,
		CheckInDb:           c.CheckInDb,
		AddCommand:          c.AddCommand,
		RemoveCustom:        c.RemoveCustom,
		RemoveFromDB:        c.RemoveFromDB,
		GetCustomCommands:   c.GetCustomCommands,
		Close:               c.Close,
		AddToBlacklist:      c.AddToBlacklist,
		RemoveFromBlacklist: c.RemoveFromBlacklist,
	}
}

// AddToDb adds a song to the db, so next time we encounter it we don't need to call yt-dlp
func addToDb(el queue.Element, exist bool) {
	// We check for empty strings, just to be sure
	if el.Link != "" && el.ID != "" && el.Title != "" && el.Duration != "" {
		if !exist {
			_, err := db.Exec("INSERT IGNORE INTO song (id, title, duration, thumbnail, segments) VALUES (?, ?, ?, ?, ?)",
				el.ID, el.Title, el.Duration, el.Thumbnail, database.EncodeSegments(el.Segments))
			if err != nil {
				lit.Error("Error inserting into song, %s", err)
			}
		}

		_, err := db.Exec("INSERT IGNORE INTO link (songID, link) VALUES (?, ?)", el.ID, el.Link)
		if err != nil {
			lit.Error("Error inserting into link, %s", err)
		}
	}
}
