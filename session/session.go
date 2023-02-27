package session

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

const (
	SPOTIFY_TOKEN_SESSION_KEY     = "spotify-token"
	GOOGLE_USER_TOKEN_SESSION_KEY = "google-user"
	SESSION_NAME                  = "token-session"
)

type SessionManager struct {
	store *sessions.CookieStore
}

// TODO use a proper cookie store
func New() SessionManager {
	return SessionManager{store: sessions.NewCookieStore([]byte("1234"))}
}

func (s SessionManager) RefreshToken(
	providerName string,
	r *http.Request,
	w http.ResponseWriter) (*oauth2.Token, error) {

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	existingTokens, err := s.GetSessionTokens(providerName, r)
	if err != nil {
		return nil, err
	}
	newTokens, err := provider.RefreshToken(existingTokens.RefreshToken)
	if err != nil {
		return nil, err
	}
	err = s.UpdateTokens(providerName, newTokens, r, w)
	if err != nil {
		return nil, err
	}
	return newTokens, nil
}

func (s SessionManager) GetSessionTokens(provider string, r *http.Request) (*oauth2.Token, error) {
	session, err := s.store.Get(r, SESSION_NAME)
	if err != nil {
		return nil, err
	}
	if provider == "google" {
		return session.Values[GOOGLE_USER_TOKEN_SESSION_KEY].(*oauth2.Token), nil
	} else if provider == "spotify" {
		return session.Values[SPOTIFY_TOKEN_SESSION_KEY].(*oauth2.Token), nil
	} else {
		return nil, errors.New(fmt.Sprintf("invalid provider %s", provider))
	}
}

func (s SessionManager) UpdateTokens(provider string, tokens *oauth2.Token, r *http.Request, w http.ResponseWriter) error {
	session, err := s.store.Get(r, SESSION_NAME)
	if err != nil {
		return err
	}
	if provider == "spotify" {
		session.Values[SPOTIFY_TOKEN_SESSION_KEY] = tokens
	} else if provider == "google" {
		session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = tokens
	} else {
		return errors.New(fmt.Sprintf("invalid provider %s", provider))
	}
	return session.Save(r, w)
}
