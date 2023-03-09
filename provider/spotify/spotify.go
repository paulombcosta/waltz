package spotify

import (
	"context"
	"errors"
	"log"

	"github.com/paulombcosta/waltz/provider"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
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

func (s SpotifyProvider) CreatePlaylist(name string) (provider.PlaylistID, error) {
	return "", errors.New("not implemented")
}

func (s SpotifyProvider) FindTrack(name string) (provider.TrackID, error) {
	return "", errors.New("not implemented")
}

func (s SpotifyProvider) FindPlaylistByName(name string) (provider.PlaylistID, error) {
	return "", errors.New("not implemented")
}

func (s SpotifyProvider) AddToPlaylist(playlistId string, tracks []provider.Track) error {
	return errors.New("not implemented")
}

func (s SpotifyProvider) GetFullPlaylist(id string) (*provider.FullPlaylist, error) {
	client, err := s.getSpotifyClient()
	if err != nil {
		return nil, err
	}
	fullPlaylist, err := client.GetPlaylist(context.Background(), spotify.ID(id))
	if err != nil {
		return nil, err
	}
	trackPage := fullPlaylist.Tracks
	log.Println("number of tracks: ", trackPage.Total)
	tracks := []provider.Track{}
	for _, t := range trackPage.Tracks {
		artists := []string{}
		for _, a := range t.Track.Artists {
			artists = append(artists, a.Name)
		}
		tracks = append(tracks, provider.Track{
			Name:    t.Track.Name,
			Artists: artists,
		})
	}
	return &provider.FullPlaylist{
		Tracks: tracks,
	}, nil
}

func (s SpotifyProvider) FindPlayListById(id string) (*provider.Playlist, error) {
	client, err := s.getSpotifyClient()
	if err != nil {
		return nil, err
	}
	p, err := client.GetPlaylist(context.Background(), spotify.ID(id))
	if err != nil {
		return nil, err
	}
	return &provider.Playlist{
		ID:   provider.PlaylistID(p.ID.String()),
		Name: p.Name,
	}, nil
}

func (s SpotifyProvider) GetPlaylists() ([]provider.Playlist, error) {
	client, err := s.getSpotifyClient()
	if err != nil {
		return nil, err
	}
	playlists := []provider.Playlist{}
	offset := 0
	var page *spotify.SimplePlaylistPage
	for {
		page, err = getPaginatedPlaylists(client, context.Background(), offset)
		if err != nil {
			return nil, err
		}
		for _, p := range page.Playlists {
			playlists = append(playlists, provider.Playlist{
				ID:      provider.PlaylistID(p.ID.String()),
				Name:    p.Name,
				Tracks:  p.Tracks.Total,
				Creator: p.Owner.DisplayName,
			})
		}
		if page.Next == "" || len(page.Playlists) == 0 {
			break
		}
		offset = offset + len(page.Playlists)
	}
	return playlists, nil
}

func getPaginatedPlaylists(client *spotify.Client, ctx context.Context, offset int) (*spotify.SimplePlaylistPage, error) {
	if offset == 0 {
		return client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50))
	} else {
		return client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50), spotify.Offset(offset))
	}
}

func (s SpotifyProvider) getSpotifyClient() (*spotify.Client, error) {
	token, err := s.tokenProvider.GetToken()
	if err != nil {
		return nil, err
	}
	if token != nil {
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
