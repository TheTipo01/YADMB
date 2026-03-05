package main

import "github.com/disgoorg/snowflake/v2"

// Config holds data parsed from the config.yml
type Config struct {
	Token        string         `fig:"token" validate:"required"`
	Owner        []snowflake.ID `fig:"owner" validate:"required"`
	ClientID     string         `fig:"clientid"`
	ClientSecret string         `fig:"clientsecret"`
	DSN          string         `fig:"datasourcename" validate:"required"`
	Driver       string         `fig:"drivername" validate:"required"`
	LogLevel     string         `fig:"loglevel" validate:"required"`
	YouTubeAPI   string         `fig:"youtubeapikey"`
	Address      string         `fig:"address"`
	Origin       string         `fig:"origin"`
	ApiTokens    []apiToken     `fig:"apitokens"`
	WhiteList    bool           `fig:"whitelist"`
	GuildList    []snowflake.ID `fig:"guildlist"`
}

type apiToken struct {
	UserID      snowflake.ID `fig:"userid" validate:"required"`
	Token       string       `fig:"token" validate:"required"`
	TextChannel snowflake.ID `fig:"textchannelid" validate:"required"`
	Guild       snowflake.ID `fig:"guildid" validate:"required"`
}
