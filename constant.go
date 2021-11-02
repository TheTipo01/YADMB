package main

// Used to store the various messages that we send and the directory and extension for the audio cache

const (
	cachePath      = "./audio_cache/"
	audioExtension = ".dca"
)

// Titles for embeds
const (
	enqueuedTitle     = "Enqueued"
	errorTitle        = "Error"
	skipTitle         = "Skipped"
	queueTitle        = "Queue"
	pauseTitle        = "Pause"
	lyricsTitle       = "Lyrics"
	summonTitle       = "Summon"
	disconnectedTitle = "Disconnected"
	gotoTitle         = "Goto"
	statsTitle        = "Stats™"
	restartTitle      = "Restart"
	successfulTitle   = "Successful"
	commandsTitle     = "Commands"
)

// Messages for embeds
const (
	// Voice channel
	notInVC    = "You're not in a voice channel in this guild!"
	cantJoinVC = "Can't join voice channel!\n"

	// Queue
	queueCleared = "Queue cleared!"
	queueEmpty   = "Queue is empty!"

	// Song status
	paused    = "Paused the current song"
	resumed   = "Resumed the current song"
	searching = "Searching..."
	skippedTo = "Skipped to "

	// Custom commands
	commandAdded   = "Custom command added!"
	commandRemoved = "Custom command removed!"
	commandInvalid = "Not a valid custom command!\nSee /listcustom for a list of custom commands."

	// Errors
	notCached       = "Song is not cached!"
	errorNotPlaying = "To use this command something needs to be playing!"
	invalidURL      = "Invalid URL!"
	stillPlaying    = "Can't disconnect the bot!\nStill playing in a voice channel."
	gotoInvalid     = "Wrong format.\nValid formats are: 1h10m3s, 3m, 4m10s..."
	nothingPlaying  = "No songs playing!"
	spotifyError    = "Can't get info about spotify playlist!\nError code: "
	nothingFound    = "No song found!\n"

	// Feedback
	disconnected = "Bye-bye!"
	joined       = "Joined "
)
