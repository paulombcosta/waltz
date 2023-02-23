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

func (y SpotifyProvider) IsLoggedIn() bool {
	_, err := y.tokenProvider.RefreshToken()
	return err == nil
}

func (y SpotifyProvider) GetPlaylists() ([]provider.Playlist, error) {
	// TODO
	return nil, nil
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
