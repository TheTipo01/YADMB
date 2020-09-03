package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/url"
	"time"
)

const (
	tblSong     = "CREATE TABLE IF NOT EXISTS `song` (`link` varchar(500) NOT NULL, `id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, PRIMARY KEY (`link`))"
	tblCommands = "CREATE TABLE IF NOT EXISTS `customCommands` (`id` int(11) NOT NULL AUTO_INCREMENT, `guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL, PRIMARY KEY (`id`), UNIQUE KEY `command` (`command`,`song`))"
)

//Logs and instantly delete a message
func deleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Println(m.Author.Username + ": " + m.Content)
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		fmt.Println("Can't delete message,", err)
	}
}

//Finds user current voice channel
func findUserVoiceState(session *discordgo.Session, m *discordgo.MessageCreate) string {
	user := m.Author.ID

	//TODO: Better webhook handling
	//My user id, for playing song via a webhook
	if m.WebhookID != "" {
		user = "145618075452964864"
	}

	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == user {
				return vs.ChannelID
			}
		}
	}

	return ""
}

//Checks if a string is a valid URL
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

//Removes element from the queue
func removeFromQueue(id string, guild string) {
	for i, q := range queue[guild] {
		if q.id == id {
			copy(queue[guild][i:], queue[guild][i+1:])
			queue[guild][len(queue[guild])-1] = Queue{"", "", "", "", "", nil, 0, ""}
			queue[guild] = queue[guild][:len(queue[guild])-1]
			return
		}
	}
}

//Sends and delete after three second an embed in a given channel
func sendAndDeleteEmbed(s *discordgo.Session, embed *discordgo.MessageEmbed, txtChannel string) {
	m, err := s.ChannelMessageSendEmbed(txtChannel, embed)
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 3)

	err = s.ChannelMessageDelete(txtChannel, m.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
}

//Finds pointer for a given song id
func findQueuePointer(guildId, id string) int {
	for i := range queue[guildId] {
		if queue[guildId][i].id == id {
			return i
		}
	}

	return -1
}

//Formats a string given it's duration in seconds
func formatDuration(duration float64) string {
	duration2 := int(duration)
	hours := duration2 / 3600
	duration2 = duration2 - 3600*hours
	minutes := (duration2) / 60
	duration2 = duration2 - minutes*60

	if hours != 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, duration2)
	} else {
		if minutes != 0 {
			return fmt.Sprintf("%02d:%02d", minutes, duration2)
		} else {
			return fmt.Sprintf("%02d", duration2)
		}
	}
}

//Executes a simple query given a DB
func execQuery(query string, db *sql.DB) {
	statement, err := db.Prepare(query)
	if err != nil {
		log.Println("Error preparing query,", err)
		return
	}

	_, err = statement.Exec()
	if err != nil {
		log.Println("Error creating table,", err)
	}
}

//Adds a song to the db, so next time we encounter it we don't need to call youtube-dl
func addToDb(el Queue) {
	statement, _ := db.Prepare("INSERT INTO song (link, id, title, duration) VALUES(?, ?, ?, ?)")

	_, err := statement.Exec(el.link, el.id, el.title, el.duration)
	if err != nil {
		log.Println("Error inserting into the database,", err)
	}
}

//Checks if we already have downloaded a song and we've got info about it
func checkInDb(link string) Queue {
	var el Queue
	el.link = link
	row := db.QueryRow("SELECT * FROM song WHERE link = ?", link)
	_ = row.Scan(&el.link, &el.id, &el.title, &el.duration)

	return el
}

//Adds a custom command to db and to the command map
func addCommand(command string, song string, guild string) {
	statement, _ := db.Prepare("INSERT INTO customCommands (guild, command, song) VALUES(?, ?, ?)")

	_, err := statement.Exec(guild, command, song)
	if err != nil {
		log.Println("Error inserting into the database,", err)
	}

	custom[guild] = append(custom[guild], CustomCommand{command, song})
}

//Loads custom command from the database
func loadCustomCommands(db *sql.DB) {
	var id int
	var guild, command, song string

	rows, err := db.Query("SELECT * FROM customCommands")
	if err != nil {
		log.Println("Error querying database,", err)
	}

	for rows.Next() {
		err = rows.Scan(&id, &guild, &command, &song)
		if err != nil {
			log.Println("Error scanning rows from query,", err)
			continue
		}

		custom[guild] = append(custom[guild], CustomCommand{command, song})
	}
}
