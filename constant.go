package main

// This file is used to hold all things constant from the bot

const (
	// How many DCA frames are needed for a second. It's not perfect, but good enough.
	frameSeconds = 50.00067787
)

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
	disconnectedTitle = "Disconnected"
	gotoTitle         = "Goto"
	statsTitle        = "Stats™"
	restartTitle      = "Restart"
	successfulTitle   = "Successful"
	commandsTitle     = "Commands"
	blacklistTitle    = "Blacklist"
	resumeTitle       = "Resume"
	djTitle           = "DJ"
)

// Messages for embeds
const (
	// Voice channel
	notInVC    = "You're not in a voice channel in this guild!"
	cantJoinVC = "Can't join voice channel!"

	// Queue
	queueCleared = "Queue cleared!"
	queueEmpty   = "Queue is empty!"

	// Song status
	paused         = "Paused the current song"
	alreadyPaused  = "The song is already paused"
	resumed        = "Resumed the current song"
	alreadyResumed = "The song is already playing"
	skippedTo      = "Skipped to "

	// Custom commands
	commandAdded   = "Custom command added!"
	commandRemoved = "Custom command removed!"
	commandInvalid = "Not a valid custom command!\nSee /listcustom for a list of custom commands."

	// Errors
	notCached           = "Song is not cached!"
	invalidURL          = "Invalid URL!"
	stillPlaying        = "Can't disconnect the bot!\nStill playing in a voice channel."
	gotoInvalid         = "Wrong format.\nValid formats are: 1h10m3s, 3m, 4m10s..."
	nothingPlaying      = "No song playing!"
	spotifyError        = "Can't get info about spotify link!\nError code: "
	spotifyNotConfigure = "Spotify is not configured!\nSee the documentation for more info."
	commandExists       = "Command already exists!"
	commandNotExists    = "Command doesn't exist!"

	// Feedback
	disconnected = "Bye-bye!"

	// DJ
	djEnabled     = "DJ mode enabled!"
	djDisabled    = "DJ mode disabled!"
	djNot         = "User is not a DJ, and DJ mode is enabled!"
	djRoleChanged = "DJ role changed!"
	djRoleEqual   = "DJ role is already that role!"
)
