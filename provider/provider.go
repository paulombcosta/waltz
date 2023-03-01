package provider

import (
	"golang.org/x/oauth2"
)

type PlaylistID string

type TokenProvider interface {
	GetToken() (*oauth2.Token, error)
	RefreshToken() (*oauth2.Token, error)
}

type Provider interface {
	IsLoggedIn() bool
	GetPlaylists() ([]Playlist, error)
	CreatePlaylist(name string) (*PlaylistID, error)
}

type Playlist struct {
	ID      PlaylistID
	Name    string
	Tracks  uint
	Creator string
}
