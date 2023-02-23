package provider

import (
	"net/http"

	"golang.org/x/oauth2"
)

type TokenProvider interface {
	GetToken() (oauth2.Token, error)
	RefreshToken(oauth2.Token) error
}

// have a token provider instead of having use these http types
type Provider interface {
	IsLoggedIn(r *http.Request, w http.ResponseWriter) bool
	GetPlaylists(r *http.Request) ([]Playlist, error)
}

type Playlist struct {
	Name string
}
