package provider

import (
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

type PlaylistID string
type TrackID string

type TokenProvider interface {
	GetToken() (*oauth2.Token, error)
	RefreshToken() (*oauth2.Token, error)
}

//go:generate mockery --name Provider
type Provider interface {
	Name() string
	IsLoggedIn() bool
	GetPlaylists() ([]Playlist, error)
	CreatePlaylist(name string) (PlaylistID, error)
	FindTrack(name string) (TrackID, error)
	FindPlaylistByName(name string) (PlaylistID, error)
	GetFullPlaylist(id string) (*FullPlaylist, error)
	AddToPlaylist(playlistId string, trackId string) error
}

type FullPlaylist struct {
	Playlist
	Tracks []Track
}

type Track struct {
	ID      string
	Name    string
	Artists []string
}

func (t Track) FullName() string {
	artists := strings.Join(t.Artists, ", ")
	return fmt.Sprintf("%s - %s", artists, t.Name)
}

type Playlist struct {
	ID      PlaylistID
	Name    string
	Tracks  uint
	Creator string
}
