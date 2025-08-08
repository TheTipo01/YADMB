package manager

import (
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TheTipo01/YADMB/api/notification"
	"github.com/TheTipo01/YADMB/constants"
	"github.com/TheTipo01/YADMB/queue"
	"github.com/bwmarrin/lit"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// filterPlaylist checks if the link is from YouTube: if yes, it removes the playlist parameter.
// And if it doesn't contain an ID for the video, it returns an error.
func filterPlaylist(link string) (string, error) {
	if com, be := strings.Contains(link, "youtube.com"), strings.Contains(link, "youtu.be"); com || be {
		u, err := url.Parse(link)
		if err != nil {
			lit.Error("filterPlaylist failed: %s", err)
			return "", err
		}

		q := u.Query()
		q.Del("list")
		q.Del("index")
		if q.Has("v") || be {
			u.RawQuery = q.Encode()
			return u.String(), nil
		}

		// Shorts link don't have a parameter for the video ID
		if !strings.Contains(link, "shorts") {
			return "", errors.New("no video ID found")
		}
	}

	return link, nil
}

// IsValidURL Checks if a string is a valid URL
func IsValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	return err == nil
}

// FormatDuration Formats a string given its duration in seconds
func FormatDuration(duration float64) string {
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

// Shuffles a slice of strings
func shuffle(a []string) []string {
	final := make([]string, len(a))

	for i, v := range rand.Perm(len(a)) {
		final[v] = a[i]
	}
	return final
}

// FolderStats gets the size of a directory and the number of files in it
func FolderStats(path string) (size int64, i int) {
	symlink, _ := filepath.EvalSymlinks(path)
	_ = filepath.Walk(symlink, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
			i++
		}
		return err
	})

	return size, i
}

// ByteCountSI formats bytes into a readable format
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func DeleteInteraction(client bot.Client, i discord.Interaction, c <-chan struct{}) {
	if c != nil {
		<-c
	}

	err := client.Rest().DeleteInteractionResponse(i.ApplicationID(), i.Token())
	if err != nil {
		lit.Error("InteractionResponseDelete failed: %s", err)
		return
	}
}

// idGen returns the first 11 characters of the SHA1 hash for the given link
func idGen(link string) string {
	h := sha1.New()
	h.Write([]byte(link))

	return strings.ToLower(base32.HexEncoding.EncodeToString(h.Sum(nil))[0:11])
}

func checkAudioOnly(formats RequestedFormats) bool {
	for _, f := range formats {
		if f.Resolution == "audio only" {
			return true
		}
	}

	return false
}

// IsCommandNotAvailable isCommandNotAvailable checks whatever a command is available
func IsCommandNotAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err != nil
}

func HasRole(roles []snowflake.ID, role string) bool {
	for _, r := range roles {
		if r.String() == role {
			return true
		}
	}
	return false
}

// cleanURL removes tracking and other unnecessary parameters from a URL
func cleanURL(link string) string {
	u, _ := url.Parse(link)
	q := u.Query()

	q.Del("utm_source")
	q.Del("feature")
	q.Del("si")

	u.RawQuery = q.Encode()

	return u.String()
}

func notify(n notification.NotificationMessage) {
	Notifications <- n
}

// skipTo checks if the link has a time parameter and if so, skips to that time
func skipTo(link string, el *queue.Element) {
	u, _ := url.Parse(link)
	q := u.Query()
	checkTimeParameter(q, el)
}

func checkTimeParameter(q url.Values, el *queue.Element) {
	if t := q.Get("t"); t != "" {
		if number, err := strconv.Atoi(t); err == nil {
			if el.Segments == nil {
				el.Segments = make(map[int]struct{}, 2)
			}

			el.Segments[0] = struct{}{}
			el.Segments[int(float64(number)*constants.FrameSeconds)] = struct{}{}
		}
	}
}
