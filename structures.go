package main

// Config holds data parsed from the config.yml
type Config struct {
	Token        string     `fig:"token" validate:"required"`
	Owner        string     `fig:"owner" validate:"required"`
	ClientID     string     `fig:"clientid"`
	ClientSecret string     `fig:"clientsecret"`
	DSN          string     `fig:"datasourcename" validate:"required"`
	Driver       string     `fig:"drivername" validate:"required"`
	LogLevel     string     `fig:"loglevel" validate:"required"`
	YouTubeAPI   string     `fig:"youtubeapikey"`
	Address      string     `fig:"address"`
	ApiTokens    []apiToken `fig:"apitokens"`
}

type apiToken struct {
	UserID      string `fig:"userid" validate:"required"`
	Token       string `fig:"token" validate:"required"`
	TextChannel string `fig:"textchannelid" validate:"required"`
	Guild       string `fig:"guildid" validate:"required"`
}
