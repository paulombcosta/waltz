package spotify

import (
	"context"

	"github.com/paulombcosta/waltz/provider"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyProvider struct {
	tokenProvider provider.TokenProvider
}

func New(tokenProvider provider.TokenProvider) *SpotifyProvider {
	return &SpotifyProvider{tokenProvider: tokenProvider}
}

func (s SpotifyProvider) IsLoggedIn() bool {
	_, err := s.tokenProvider.RefreshToken()
	return err == nil
}

func (s SpotifyProvider) GetPlaylists() ([]provider.Playlist, error) {
	token, err := s.tokenProvider.GetToken()
	if err != nil {
		return nil, err
	}
	client, err := s.getSpotifyClient(token)
	if err != nil {
		return nil, err
	}
	page, err := client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50))
	if err != nil {
		return nil, err
	}
	playlists := []provider.Playlist{}
	for _, p := range page.Playlists {
		playlists = append(playlists, provider.Playlist{Name: p.Name})
	}
	return playlists, nil
}

func (s SpotifyProvider) getSpotifyClient(tok *oauth2.Token) (*spotify.Client, error) {
	if tok != nil {
		newTokens, err := s.tokenProvider.RefreshToken()
		if err != nil {
			return nil, err
		}
		client := spotify.New(spotifyauth.New().Client(context.Background(), newTokens))
		return client, nil
	} else {
		return nil, nil
	}
}
