package main

type YoutubeDL struct {
	License           interface{} `json:"license"`
	AltTitle          interface{} `json:"alt_title"`
	Vbr               interface{} `json:"vbr"`
	AutomaticCaptions struct {
	} `json:"automatic_captions"`
	StartTime     interface{} `json:"start_time"`
	Uploader      string      `json:"uploader"`
	Duration      int         `json:"duration"`
	AverageRating float64     `json:"average_rating"`
	Creator       interface{} `json:"creator"`
	Subtitles     struct {
	} `json:"subtitles"`
	Thumbnail        string      `json:"thumbnail"`
	StretchedRatio   interface{} `json:"stretched_ratio"`
	DislikeCount     interface{} `json:"dislike_count"`
	FormatID         string      `json:"format_id"`
	Track            interface{} `json:"track"`
	Chapters         interface{} `json:"chapters"`
	ReleaseYear      interface{} `json:"release_year"`
	Extractor        string      `json:"extractor"`
	ID               string      `json:"id"`
	RequestedFormats []struct {
		FormatNote        string `json:"format_note"`
		Height            int    `json:"height"`
		Vcodec            string `json:"vcodec"`
		Acodec            string `json:"acodec"`
		Ext               string `json:"ext"`
		DownloaderOptions struct {
			HTTPChunkSize int `json:"http_chunk_size"`
		} `json:"downloader_options"`
		HTTPHeaders struct {
			AcceptEncoding string `json:"Accept-Encoding"`
			AcceptLanguage string `json:"Accept-Language"`
			UserAgent      string `json:"User-Agent"`
			AcceptCharset  string `json:"Accept-Charset"`
			Accept         string `json:"Accept"`
		} `json:"http_headers"`
		Fps       int         `json:"fps"`
		Width     int         `json:"width"`
		PlayerURL interface{} `json:"player_url"`
		Format    string      `json:"format"`
		FormatID  string      `json:"format_id"`
		Asr       interface{} `json:"asr"`
		Tbr       float64     `json:"tbr"`
		Protocol  string      `json:"protocol"`
		URL       string      `json:"url"`
		Filesize  int         `json:"filesize"`
		Abr       int         `json:"abr,omitempty"`
	} `json:"requested_formats"`
	WebpageURL  string      `json:"webpage_url"`
	LikeCount   interface{} `json:"like_count"`
	UploadDate  string      `json:"upload_date"`
	Album       interface{} `json:"album"`
	Annotations interface{} `json:"annotations"`
	Formats     []struct {
		Vcodec            string `json:"vcodec"`
		FormatID          string `json:"format_id"`
		Abr               int    `json:"abr,omitempty"`
		Asr               int    `json:"asr"`
		FormatNote        string `json:"format_note"`
		Ext               string `json:"ext"`
		DownloaderOptions struct {
			HTTPChunkSize int `json:"http_chunk_size"`
		} `json:"downloader_options,omitempty"`
		HTTPHeaders struct {
			AcceptEncoding string `json:"Accept-Encoding"`
			AcceptLanguage string `json:"Accept-Language"`
			UserAgent      string `json:"User-Agent"`
			AcceptCharset  string `json:"Accept-Charset"`
			Accept         string `json:"Accept"`
		} `json:"http_headers"`
		Fps       interface{} `json:"fps"`
		Width     interface{} `json:"width"`
		PlayerURL interface{} `json:"player_url"`
		Acodec    string      `json:"acodec"`
		Format    string      `json:"format"`
		Height    interface{} `json:"height"`
		Tbr       float64     `json:"tbr"`
		Filesize  int         `json:"filesize"`
		Protocol  string      `json:"protocol"`
		URL       string      `json:"url"`
		Container string      `json:"container,omitempty"`
	} `json:"formats"`
	WebpageURLBasename string      `json:"webpage_url_basename"`
	PlaylistIndex      interface{} `json:"playlist_index"`
	DisplayID          string      `json:"display_id"`
	ViewCount          int         `json:"view_count"`
	Categories         []string    `json:"categories"`
	EndTime            interface{} `json:"end_time"`
	RequestedSubtitles interface{} `json:"requested_subtitles"`
	ReleaseDate        interface{} `json:"release_date"`
	Fulltitle          string      `json:"fulltitle"`
	Title              string      `json:"title"`
	Height             int         `json:"height"`
	ChannelURL         string      `json:"channel_url"`
	Filename           string      `json:"_filename"`
	Width              int         `json:"width"`
	ChannelID          string      `json:"channel_id"`
	Thumbnails         []struct {
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		Resolution string `json:"resolution"`
		URL        string `json:"url"`
		ID         string `json:"id"`
	} `json:"thumbnails"`
	Tags          []string    `json:"tags"`
	UploaderURL   string      `json:"uploader_url"`
	Artist        interface{} `json:"artist"`
	Abr           int         `json:"abr"`
	Vcodec        string      `json:"vcodec"`
	Acodec        string      `json:"acodec"`
	Ext           string      `json:"ext"`
	AgeLimit      int         `json:"age_limit"`
	Fps           int         `json:"fps"`
	UploaderID    string      `json:"uploader_id"`
	IsLive        interface{} `json:"is_live"`
	Format        string      `json:"format"`
	Description   string      `json:"description"`
	EpisodeNumber interface{} `json:"episode_number"`
	SeasonNumber  interface{} `json:"season_number"`
	Resolution    interface{} `json:"resolution"`
	Playlist      interface{} `json:"playlist"`
	ExtractorKey  string      `json:"extractor_key"`
	Series        interface{} `json:"series"`
}
