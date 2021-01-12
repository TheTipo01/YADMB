package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	tblSong     = "CREATE TABLE IF NOT EXISTS `song` (`link` varchar(500) NOT NULL, `id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, PRIMARY KEY (`link`))"
	tblCommands = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL,  PRIMARY KEY (`guild`,`command`,`song`))"
)

// Logs and instantly delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	lit.Debug(m.Author.Username + ": " + m.Content)

	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		lit.Error("Can't delete message, %s", err)
	}
}

// Finds user current voice channel
func findUserVoiceState(session *discordgo.Session, m *discordgo.MessageCreate) *discordgo.VoiceState {

	for _, guild := range session.State.Guilds {
		if guild.ID != m.GuildID {
			continue
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == m.Author.ID {
				return vs
			}
		}
	}

	return nil
}

// Checks if a string is a valid URL
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

// Removes element from the queue
func removeFromQueue(id string, guild string) {
	for i, q := range server[guild].queue {
		if q.id == id {
			copy(server[guild].queue[i:], server[guild].queue[i+1:])
			server[guild].queue[len(server[guild].queue)-1] = Queue{"", "", "", "", "", nil, 0, "", nil}
			server[guild].queue = server[guild].queue[:len(server[guild].queue)-1]
			return
		}
	}
}

// Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		lit.Error("MessageSendEmbed failed: %s", err)
		return
	}

	time.Sleep(time.Second * 5)

	err = s.ChannelMessageDelete(txtChannel, m.ID)
	if err != nil {
		lit.Error("MessageDelete failed: %s", err)
		return
	}
}

// Formats a string given it's duration in seconds
func formatDuration(duration float64) string {
	duration2 := int(duration)
	hours := duration2 / 3600
	duration2 -= 3600 * hours
	minutes := (duration2) / 60
	duration2 -= minutes * 60

	if hours != 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, duration2)
	}

	if minutes != 0 {
		return fmt.Sprintf("%02d:%02d", minutes, duration2)
	}

	return fmt.Sprintf("%02d", duration2)

}

// Executes a simple query given a DB
func execQuery(query string, db *sql.DB) {
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
		statement, _ := db.Prepare("INSERT INTO song (link, id, title, duration) VALUES(?, ?, ?, ?)")

		_, err := statement.Exec(el.link, el.id, el.title, el.duration)
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
	_ = row.Scan(&el.link, &el.id, &el.title, &el.duration)

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
func loadCustomCommands(db *sql.DB) {
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

// Split lyrics into smaller messages
func formatLongMessage(text []string) []string {
	var counter int
	var output []string
	var buffer string
	const charLimit = 1900

	for _, line := range text {
		counter += strings.Count(line, "")

		// If the counter is exceeded, we append all the current line to the final slice
		if counter > charLimit {
			counter = 0
			output = append(output, buffer)

			buffer = line + "\n"
			continue
		}

		buffer += line + "\n"

	}

	return append(output, buffer)
}

// Deletes an array of discordgo.Message
func deleteMessages(s *discordgo.Session, messages []discordgo.Message) {
	for _, m := range messages {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

// Shuffles a slice of strings
func shuffle(a []string) []string {
	final := make([]string, len(a))

	for i, v := range rand.Perm(len(a)) {
		final[v] = a[i]
	}
	return final
}

// Disconnects the bot from the voice channel after 1 minute if nothing is playing
func quitVC(guildID string) {
	time.Sleep(1 * time.Minute)

	if len(server[guildID].queue) == 0 && server[guildID].vc != nil {
		server[guildID].server.Lock()

		_ = server[guildID].vc.Disconnect()
		server[guildID].vc = nil

		server[guildID].server.Unlock()
	}
}

// Wrapper function for playing songs
func play(s *discordgo.Session, song, textChannel, voiceChannel, guild, username string, random bool) {
	switch {
	case strings.HasPrefix(song, "spotify:playlist:"):
		spotifyPlaylist(s, guild, voiceChannel, username, song, textChannel, random)
		break

	case isValidURL(song):
		downloadAndPlay(s, guild, voiceChannel, song, username, textChannel, random)
		break

	default:
		searchDownloadAndPlay(s, guild, voiceChannel, song, username, textChannel, random)
	}
}
