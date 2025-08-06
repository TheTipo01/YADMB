package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type Spotify struct {
	// Spotify client
	client *spotify.Client
	// Spotify context
	ctx context.Context
	// Spotify clientID
	clientID string
	// Spotify clientSecret
	clientSecret string
}

func NewSpotify(clientID, clientSecret string) (*Spotify, error) {
	s := &Spotify{
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	if err := s.refreshClient(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Spotify) refreshClient() error {
	// Spotify credentials
	credentials := clientcredentials.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	s.ctx = context.Background()

	// Check spotify token and create a spotify client
	token, err := credentials.Token(s.ctx)
	if err != nil {
		return err
	}

	s.client = spotify.New(spotifyauth.New().Client(s.ctx, token))

	return nil
}

func (s *Spotify) GetPlaylist(id spotify.ID) (*spotify.FullPlaylist, error) {
	p, err := s.client.GetPlaylist(s.ctx, id)
	if err != nil {
		err = s.refreshClient()
		if err != nil {
			return nil, err
		}

		return s.client.GetPlaylist(s.ctx, id)
	}

	return p, nil
}

func (s *Spotify) GetTrack(id spotify.ID) (*spotify.FullTrack, error) {
	t, err := s.client.GetTrack(s.ctx, id)
	if err != nil {
		err = s.refreshClient()
		if err != nil {
			return nil, err
		}

		return s.client.GetTrack(s.ctx, id)
	}

	return t, nil
}

func (s *Spotify) GetAlbum(id spotify.ID) (*spotify.FullAlbum, error) {
	a, err := s.client.GetAlbum(s.ctx, id)
	if err != nil {
		err = s.refreshClient()
		if err != nil {
			return nil, err
		}

		return s.client.GetAlbum(s.ctx, id)
	}

	return a, nil
}
