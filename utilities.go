package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

const (
	tblSong     = "CREATE TABLE IF NOT EXISTS `song` (`link` varchar(500) NOT NULL, `id` varchar(200) NOT NULL, `title` varchar(200) NOT NULL, `duration` varchar(20) NOT NULL, PRIMARY KEY (`link`))"
	tblCommands = "CREATE TABLE IF NOT EXISTS `customCommands` (`guild` varchar(18) NOT NULL, `command` varchar(100) NOT NULL, `song` varchar(100) NOT NULL,  PRIMARY KEY (`guild`,`command`,`song`))"
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
			queue[guild][len(queue[guild])-1] = Queue{"", "", "", "", "", nil, 0, "", nil}
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
	//We check for empty strings, just to be sure
	if el.link != "" && el.id != "" && el.title != "" && el.duration != "" {
		statement, _ := db.Prepare("INSERT INTO song (link, id, title, duration) VALUES(?, ?, ?, ?)")

		_, err := statement.Exec(el.link, el.id, el.title, el.duration)
		if err != nil {
			log.Println("Error inserting into the database,", err)
		}
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
	//If the song is already in the map, we ignore it
	if custom[guild][command] == song {
		return
	}

	//Else, we add it to the map
	custom[guild][command] = song

	//And to the database
	statement, _ := db.Prepare("INSERT INTO customCommands (guild, command, song) VALUES(?, ?, ?)")

	_, err := statement.Exec(guild, command, song)
	if err != nil {
		log.Println("Error inserting into the database,", err)
	}

}

//Removes a custom command from the db and from the command map
func removeCustom(command string, guild string) {
	//Remove from DB
	statement, _ := db.Prepare("DELETE FROM customCommands WHERE guild=? AND command=?")
	_, err := statement.Exec(guild, command)
	if err != nil {
		log.Println("Error removing from the database,", err)
	}

	//Remove from the map
	delete(custom[guild], command)
}

//Loads custom command from the database
func loadCustomCommands(db *sql.DB) {
	var guild, command, song string

	rows, err := db.Query("SELECT * FROM customCommands")
	if err != nil {
		log.Println("Error querying database,", err)
	}

	for rows.Next() {
		err = rows.Scan(&guild, &command, &song)
		if err != nil {
			log.Println("Error scanning rows from query,", err)
			continue
		}

		if custom[guild] == nil {
			custom[guild] = make(map[string]string)
		}

		custom[guild][command] = song
	}
}

//Split lyrics into smaller messages
func formatLongMessage(text []string) []string {
	var counter int
	var output []string
	var buffer string
	const charLimit = 1900

	for _, line := range text {
		counter += strings.Count(line, "")

		//If the counter is exceeded, we append all the current line to the final slice
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

func deleteMessages(s *discordgo.Session, messages []discordgo.Message) {
	for _, m := range messages {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

//Shuffles a slice of strings
func shuffle(a []string) []string {
	final := make([]string, len(a))

	for i, v := range rand.Perm(len(a)) {
		final[v] = a[i]
	}
	return final
}
