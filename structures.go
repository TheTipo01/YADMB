package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

type YoutubeDL struct {
	License           interface{} `json:"license"`
	AltTitle          interface{} `json:"alt_title"`
	Vbr               interface{} `json:"vbr"`
	AutomaticCaptions struct {
	} `json:"automatic_captions"`
	StartTime     interface{} `json:"start_time"`
	Uploader      string      `json:"uploader"`
	Duration      float64     `json:"duration"`
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

//Structure for holding infos about a song
type Queue struct {
	//Title of the song
	title string
	//Duration of the song
	duration string
	//ID of the song
	id string
	//Link of the song
	link string
	//User who requested the song
	user string
	//When we started playing the song
	time *time.Time
	//Offset for when we pause the song
	offset float64
	//When song is paused, we save where we were
	lastTime string
	//Message  to delete at the end of the song play
	messageID []discordgo.Message
}

//Structure for getting lyrics of a song
type Lyrics struct {
	Type                     string      `json:"_type"`
	AnnotationCount          int         `json:"annotation_count"`
	APIPath                  string      `json:"api_path"`
	FullTitle                string      `json:"full_title"`
	HeaderImageThumbnailURL  string      `json:"header_image_thumbnail_url"`
	HeaderImageURL           string      `json:"header_image_url"`
	ID                       int         `json:"id"`
	Instrumental             bool        `json:"instrumental"`
	LyricsOwnerID            int         `json:"lyrics_owner_id"`
	LyricsState              string      `json:"lyrics_state"`
	LyricsUpdatedAt          int         `json:"lyrics_updated_at"`
	Path                     string      `json:"path"`
	PyongsCount              interface{} `json:"pyongs_count"`
	SongArtImageThumbnailURL string      `json:"song_art_image_thumbnail_url"`
	SongArtImageURL          string      `json:"song_art_image_url"`
	Stats                    struct {
		AcceptedAnnotations   int  `json:"accepted_annotations"`
		Contributors          int  `json:"contributors"`
		IqEarners             int  `json:"iq_earners"`
		Transcribers          int  `json:"transcribers"`
		UnreviewedAnnotations int  `json:"unreviewed_annotations"`
		VerifiedAnnotations   int  `json:"verified_annotations"`
		Hot                   bool `json:"hot"`
	} `json:"stats"`
	Title             string `json:"title"`
	TitleWithFeatured string `json:"title_with_featured"`
	UpdatedByHumanAt  int    `json:"updated_by_human_at"`
	URL               string `json:"url"`
	PrimaryArtist     struct {
		APIPath        string `json:"api_path"`
		HeaderImageURL string `json:"header_image_url"`
		ID             int    `json:"id"`
		ImageURL       string `json:"image_url"`
		IsMemeVerified bool   `json:"is_meme_verified"`
		IsVerified     bool   `json:"is_verified"`
		Name           string `json:"name"`
		URL            string `json:"url"`
	} `json:"primary_artist"`
	AppleMusicID        interface{} `json:"apple_music_id"`
	AppleMusicPlayerURL string      `json:"apple_music_player_url"`
	Description         struct {
		Plain string `json:"plain"`
	} `json:"description"`
	EmbedContent            string      `json:"embed_content"`
	FeaturedVideo           bool        `json:"featured_video"`
	LyricsPlaceholderReason interface{} `json:"lyrics_placeholder_reason"`
	RecordingLocation       interface{} `json:"recording_location"`
	ReleaseDate             interface{} `json:"release_date"`
	ReleaseDateForDisplay   interface{} `json:"release_date_for_display"`
	CurrentUserMetadata     struct {
		Permissions         []string `json:"permissions"`
		ExcludedPermissions []string `json:"excluded_permissions"`
		Interactions        struct {
			Pyong     bool `json:"pyong"`
			Following bool `json:"following"`
		} `json:"interactions"`
		Relationships struct {
			PinnedRole interface{} `json:"pinned_role"`
		} `json:"relationships"`
		IqByAction struct {
			EditMetadata struct {
				Primary struct {
					Multiplier int     `json:"multiplier"`
					Base       float64 `json:"base"`
					Applicable bool    `json:"applicable"`
				} `json:"primary"`
			} `json:"edit_metadata"`
		} `json:"iq_by_action"`
	} `json:"current_user_metadata"`
	Album                 interface{}   `json:"album"`
	CustomPerformances    []interface{} `json:"custom_performances"`
	DescriptionAnnotation struct {
		Type           string `json:"_type"`
		AnnotatorID    int    `json:"annotator_id"`
		AnnotatorLogin string `json:"annotator_login"`
		APIPath        string `json:"api_path"`
		Classification string `json:"classification"`
		Fragment       string `json:"fragment"`
		ID             int    `json:"id"`
		IsDescription  bool   `json:"is_description"`
		Path           string `json:"path"`
		Range          struct {
			Content string `json:"content"`
		} `json:"range"`
		SongID               int           `json:"song_id"`
		URL                  string        `json:"url"`
		VerifiedAnnotatorIds []interface{} `json:"verified_annotator_ids"`
		Annotatable          struct {
			APIPath          string `json:"api_path"`
			ClientTimestamps struct {
				UpdatedByHumanAt int `json:"updated_by_human_at"`
				LyricsUpdatedAt  int `json:"lyrics_updated_at"`
			} `json:"client_timestamps"`
			Context   string `json:"context"`
			ID        int    `json:"id"`
			ImageURL  string `json:"image_url"`
			LinkTitle string `json:"link_title"`
			Title     string `json:"title"`
			Type      string `json:"type"`
			URL       string `json:"url"`
		} `json:"annotatable"`
		Annotations []struct {
			APIPath string `json:"api_path"`
			Body    struct {
				Plain string `json:"plain"`
			} `json:"body"`
			CommentCount        int         `json:"comment_count"`
			Community           bool        `json:"community"`
			CustomPreview       interface{} `json:"custom_preview"`
			HasVoters           bool        `json:"has_voters"`
			ID                  int         `json:"id"`
			Pinned              bool        `json:"pinned"`
			ShareURL            string      `json:"share_url"`
			Source              interface{} `json:"source"`
			State               string      `json:"state"`
			URL                 string      `json:"url"`
			Verified            bool        `json:"verified"`
			VotesTotal          int         `json:"votes_total"`
			CurrentUserMetadata struct {
				Permissions         []interface{} `json:"permissions"`
				ExcludedPermissions []string      `json:"excluded_permissions"`
				Interactions        struct {
					Cosign bool        `json:"cosign"`
					Pyong  bool        `json:"pyong"`
					Vote   interface{} `json:"vote"`
				} `json:"interactions"`
				IqByAction struct {
					Accept struct {
						Primary struct {
							Multiplier int     `json:"multiplier"`
							Base       float64 `json:"base"`
							Applicable bool    `json:"applicable"`
						} `json:"primary"`
					} `json:"accept"`
					Reject struct {
						Primary struct {
							Multiplier int     `json:"multiplier"`
							Base       float64 `json:"base"`
							Applicable bool    `json:"applicable"`
						} `json:"primary"`
					} `json:"reject"`
					Delete struct {
						Primary struct {
							Multiplier int     `json:"multiplier"`
							Base       float64 `json:"base"`
							Applicable bool    `json:"applicable"`
						} `json:"primary"`
					} `json:"delete"`
				} `json:"iq_by_action"`
			} `json:"current_user_metadata"`
			Authors []struct {
				Attribution float64     `json:"attribution"`
				PinnedRole  interface{} `json:"pinned_role"`
				User        struct {
					APIPath string `json:"api_path"`
					Avatar  struct {
						Tiny struct {
							URL         string `json:"url"`
							BoundingBox struct {
								Width  int `json:"width"`
								Height int `json:"height"`
							} `json:"bounding_box"`
						} `json:"tiny"`
						Thumb struct {
							URL         string `json:"url"`
							BoundingBox struct {
								Width  int `json:"width"`
								Height int `json:"height"`
							} `json:"bounding_box"`
						} `json:"thumb"`
						Small struct {
							URL         string `json:"url"`
							BoundingBox struct {
								Width  int `json:"width"`
								Height int `json:"height"`
							} `json:"bounding_box"`
						} `json:"small"`
						Medium struct {
							URL         string `json:"url"`
							BoundingBox struct {
								Width  int `json:"width"`
								Height int `json:"height"`
							} `json:"bounding_box"`
						} `json:"medium"`
					} `json:"avatar"`
					HeaderImageURL              string      `json:"header_image_url"`
					HumanReadableRoleForDisplay interface{} `json:"human_readable_role_for_display"`
					ID                          int         `json:"id"`
					Iq                          int         `json:"iq"`
					Login                       string      `json:"login"`
					Name                        string      `json:"name"`
					RoleForDisplay              interface{} `json:"role_for_display"`
					URL                         string      `json:"url"`
					CurrentUserMetadata         struct {
						Permissions         []string      `json:"permissions"`
						ExcludedPermissions []interface{} `json:"excluded_permissions"`
						Interactions        struct {
							Following bool `json:"following"`
						} `json:"interactions"`
					} `json:"current_user_metadata"`
				} `json:"user"`
			} `json:"authors"`
			CosignedBy       []interface{} `json:"cosigned_by"`
			RejectionComment interface{}   `json:"rejection_comment"`
			VerifiedBy       interface{}   `json:"verified_by"`
		} `json:"annotations"`
	} `json:"description_annotation"`
	FeaturedArtists        []interface{} `json:"featured_artists"`
	LyricsMarkedCompleteBy interface{}   `json:"lyrics_marked_complete_by"`
	Media                  []interface{} `json:"media"`
	ProducerArtists        []interface{} `json:"producer_artists"`
	SongRelationships      []struct {
		RelationshipType string        `json:"relationship_type"`
		Type             string        `json:"type"`
		Songs            []interface{} `json:"songs"`
	} `json:"song_relationships"`
	VerifiedAnnotationsBy []interface{} `json:"verified_annotations_by"`
	VerifiedContributors  []interface{} `json:"verified_contributors"`
	VerifiedLyricsBy      []interface{} `json:"verified_lyrics_by"`
	WriterArtists         []interface{} `json:"writer_artists"`
	Lyrics                string        `json:"lyrics"`
}
