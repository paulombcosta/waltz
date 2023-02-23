package provider

import (
	"golang.org/x/oauth2"
)

type TokenProvider interface {
	GetToken() (*oauth2.Token, error)
	RefreshToken() (*oauth2.Token, error)
}

// have a token provider instead of having use these http types
type Provider interface {
	IsLoggedIn() bool
	GetPlaylists() ([]Playlist, error)
}

type Playlist struct {
	Name string
}
