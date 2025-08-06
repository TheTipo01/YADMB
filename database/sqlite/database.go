package sqlite

import (
	"database/sql"

	"github.com/TheTipo01/YADMB/database"
	"github.com/TheTipo01/YADMB/database/common"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
	_ "modernc.org/sqlite"
)

const (
	tblSong      = "CREATE TABLE IF NOT EXISTS `song` (`id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, `thumbnail` varchar(500) NOT NULL, `segments` varchar(1000) NOT NULL, PRIMARY KEY (`id`));"
	tblCommands  = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL, `loop` tinyint(1) NOT NULL DEFAULT 0,  PRIMARY KEY (`guild`,`command`));"
	tblBlacklist = "CREATE TABLE IF NOT EXISTS `blacklist`(`id` VARCHAR(20) NOT NULL, PRIMARY KEY (`id`));"
	tblLink      = "create table IF NOT EXISTS link ( songID varchar(200) not null references song, link varchar(500) not null constraint link_pk primary key );"
	tblDJ        = "CREATE TABLE IF NOT EXISTS `dj` ( `guild` VARCHAR(20) NOT NULL, `role` VARCHAR(20) NULL, `enabled` TINYINT(1) NOT NULL DEFAULT '0', PRIMARY KEY (`guild`) );"
	tblFavorites = "CREATE TABLE IF NOT EXISTS `favorites`( `userID` VARCHAR(20) NOT NULL, `name` VARCHAR(100) NOT NULL, `link` VARCHAR(200) NOT NULL, `folder` VARCHAR(100) NULL DEFAULT NULL, PRIMARY KEY (`userID`, `name`));"
	tblSearch    = "create table if not exists search ( link varchar(500) not null, term text not null constraint search_pk primary key );"
	tblPlaylist  = "create table if not exists playlist ( playlist varchar(500) not null, entry varchar(500) not null, number integer not null, constraint playlist_pk primary key (playlist, entry) );"
)

var db *sql.DB

// NewDatabase returns a new database
func NewDatabase(dsn string) *database.Database {
	var err error

	// Open database connection
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		lit.Error("Error opening db connection, %s", err)
		return nil
	}

	// Create tables if they don't exist
	database.ExecQuery(db, tblSong, tblCommands, tblBlacklist, tblLink, tblDJ, tblFavorites, tblSearch, tblPlaylist)

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
		GetDJ:               c.GetDJ,
		UpdateDJRole:        updateDJRole,
		GetBlacklist:        c.GetBlacklist,
		SetDJSettings:       setDJSettings,
		AddLinkDB:           addLinkDB,
		GetFavorites:        c.GetFavorites,
		AddFavorite:         c.AddFavorite,
		RemoveFavorite:      c.RemoveFavorite,
		GetSearch:           c.GetSearch,
		AddSearch:           c.AddSearch,
		RemoveSearch:        c.RemoveSearch,
		GetPlaylist:         c.GetPlaylist,
		AddPlaylist:         c.AddPlaylist,
		RemovePlaylist:      c.RemovePlaylist,
	}
}

// AddToDb adds a song to the db, so next time we encounter it we don't need to call yt-dlp
func addToDb(el queue.Element, exist bool) {
	// We check for empty strings, just to be sure
	if el.Link != "" && el.ID != "" && el.Title != "" && el.Duration != "" {
		if !exist {
			_, err := db.Exec("INSERT OR IGNORE INTO song (id, title, duration, thumbnail, segments) VALUES (?, ?, ?, ?, ?)",
				el.ID, el.Title, el.Duration, el.Thumbnail, database.EncodeSegments(el.Segments))
			if err != nil {
				lit.Error("Error inserting into song, %s", err)
			}
		}

		err := addLinkDB(el.ID, el.Link)
		if err != nil {
			lit.Error("Error inserting into link, %s", err.Error())
		}
	}
}

func addLinkDB(id, link string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO link (songID, link) VALUES (?, ?)", id, link)
	return err
}

func setDJSettings(guild string, enabled bool) error {
	_, err := db.Exec("INSERT INTO dj (guild, enabled) VALUES (?, ?) ON CONFLICT(guild) DO UPDATE SET enabled = ?", guild, enabled, enabled)
	return err
}

func updateDJRole(guild string, role string) error {
	_, err := db.Exec("INSERT INTO dj (guild, role) VALUES (?, ?) ON CONFLICT(guild) DO UPDATE SET role = ?", guild, role, role)
	return err
}
