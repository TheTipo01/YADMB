package main

// Config holds data parsed from the config.yml
type Config struct {
	Token        string `fig:"token" validate:"required"`
	Owner        string `fig:"owner" validate:"required"`
	ClientID     string `fig:"clientid"`
	ClientSecret string `fig:"clientsecret"`
	DSN          string `fig:"datasourcename" validate:"required"`
	Driver       string `fig:"drivername" validate:"required"`
	LogLevel     string `fig:"loglevel" validate:"required"`
	YouTubeAPI   string `fig:"youtubeapikey"`
	Address      string `fig:"address"`
}
