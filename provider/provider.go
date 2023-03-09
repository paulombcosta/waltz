package provider

import (
	"golang.org/x/oauth2"
)

type PlaylistID string
type TrackID string

type TokenProvider interface {
	GetToken() (*oauth2.Token, error)
	RefreshToken() (*oauth2.Token, error)
}

type Provider interface {
	IsLoggedIn() bool
	GetPlaylists() ([]Playlist, error)
	CreatePlaylist(name string) (PlaylistID, error)
	FindTrack(name string) (TrackID, error)
	FindPlaylistByName(name string) (PlaylistID, error)
	FindPlayListById(id string) (*Playlist, error)
	GetFullPlaylist(id string) (*FullPlaylist, error)
	AddToPlaylist(playlistId string, tracks []Track) error
}

type FullPlaylist struct {
	Playlist
	Tracks []Track
}

type Track struct {
	Name string
}

type Playlist struct {
	ID      PlaylistID
	Name    string
	Tracks  uint
	Creator string
}
