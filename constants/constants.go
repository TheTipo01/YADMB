package constants

const (
	CachePath      = "./audio_cache/"
	AudioExtension = ".dca"
)

// Titles for embeds
const (
	EnqueuedTitle     = "Enqueued"
	ErrorTitle        = "Error"
	SkipTitle         = "Skipped"
	QueueTitle        = "Queue"
	PauseTitle        = "Pause"
	DisconnectedTitle = "Disconnected"
	GotoTitle         = "Goto"
	StatsTitle        = "Stats™"
	RestartTitle      = "Restart"
	SuccessfulTitle   = "Successful"
	CommandsTitle     = "Commands"
	BlacklistTitle    = "Blacklist"
	ResumeTitle       = "Resume"
	DjTitle           = "DJ"
	WebUITitle        = "Web UI"
)

// Messages for embeds
const (
	// Voice channel
	NotInVC    = "You're not in a voice channel in this guild!"
	CantJoinVC = "Can't join voice channel!"

	// Queue
	QueueCleared = "Queue cleared!"
	QueueEmpty   = "Queue is empty!"

	// Song status
	Paused         = "Paused the current song"
	AlreadyPaused  = "The song is already paused"
	Resumed        = "Resumed the current song"
	AlreadyResumed = "The song is already playing"
	SkippedTo      = "Skipped to "

	// Custom commands
	CommandAdded   = "Custom command added!"
	CommandRemoved = "Custom command removed!"
	CommandInvalid = "Not a valid custom command!\nSee /listcustom for a list of custom commands."

	// Errors
	NotCached           = "Song is not cached!"
	InvalidURL          = "Invalid URL!"
	StillPlaying        = "Can't disconnect the bot!\nStill playing in a voice channel."
	GotoInvalid         = "Wrong format.\nValid formats are: 1h10m3s, 3m, 4m10s..."
	NothingPlaying      = "No song playing!"
	SpotifyError        = "Can't get info about spotify link!\nError code: "
	SpotifyNotConfigure = "Spotify is not configured!\nSee the documentation for more info."
	CommandExists       = "Command already exists!"
	CommandNotExists    = "Command doesn't exist!"

	// Feedback
	Disconnected = "Bye-bye!"

	// DJ
	DjEnabled     = "DJ mode enabled!"
	DjDisabled    = "DJ mode disabled!"
	DjNot         = "User is not a DJ, and DJ mode is enabled!"
	DjRoleChanged = "DJ role changed!"
	DjRoleEqual   = "DJ role is already that role!"
)

const (
	// How many DCA frames are needed for a second. It's not perfect, but good enough.
	FrameSeconds = 50.00067787
)
