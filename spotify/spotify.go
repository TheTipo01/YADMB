package spotify

import (
	"context"
	"sync"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type Spotify struct {
	// Protects client and ctx, which are replaced when the token is refreshed.
	mu sync.RWMutex
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

	ctx := context.Background()

	// Check spotify token and create a spotify client
	token, err := credentials.Token(ctx)
	if err != nil {
		return err
	}

	client := spotify.New(spotifyauth.New().Client(ctx, token))

	s.mu.Lock()
	s.ctx = ctx
	s.client = client
	s.mu.Unlock()

	return nil
}

func (s *Spotify) getClient() (*spotify.Client, context.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.client, s.ctx
}

func (s *Spotify) refreshAndGetClient() (*spotify.Client, context.Context, error) {
	if err := s.refreshClient(); err != nil {
		return nil, nil, err
	}

	client, ctx := s.getClient()
	return client, ctx, nil
}

func (s *Spotify) GetPlaylist(id spotify.ID) (*spotify.FullPlaylist, error) {
	client, ctx := s.getClient()

	p, err := client.GetPlaylist(ctx, id)
	if err != nil {
		client, ctx, err = s.refreshAndGetClient()
		if err != nil {
			return nil, err
		}

		return client.GetPlaylist(ctx, id)
	}

	return p, nil
}

func (s *Spotify) GetTrack(id spotify.ID) (*spotify.FullTrack, error) {
	client, ctx := s.getClient()

	t, err := client.GetTrack(ctx, id)
	if err != nil {
		client, ctx, err = s.refreshAndGetClient()
		if err != nil {
			return nil, err
		}

		return client.GetTrack(ctx, id)
	}

	return t, nil
}

func (s *Spotify) GetAlbum(id spotify.ID) (*spotify.FullAlbum, error) {
	client, ctx := s.getClient()

	a, err := client.GetAlbum(ctx, id)
	if err != nil {
		client, ctx, err = s.refreshAndGetClient()
		if err != nil {
			return nil, err
		}

		return client.GetAlbum(ctx, id)
	}

	return a, nil
}
