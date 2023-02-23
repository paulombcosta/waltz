package main

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/markbates/goth/gothic"
	"github.com/paulombcosta/waltz/provider"
	"github.com/paulombcosta/waltz/provider/spotify"
	"github.com/paulombcosta/waltz/provider/youtube"
	"github.com/paulombcosta/waltz/token"
	"golang.org/x/oauth2"
)

const (
	PROVIDER_GOOGLE  = "google"
	PROVIDER_SPOTIFY = "spotify"
)

func (a application) getProvider(name string, r *http.Request, w http.ResponseWriter) (provider.Provider, error) {
	tokenProvider := token.New(name, r, w, a.sessionManager)
	if name == PROVIDER_GOOGLE {
		return youtube.New(tokenProvider), nil
	} else if name == PROVIDER_SPOTIFY {
		return spotify.New(tokenProvider), nil
	} else {
		return nil, errors.New("invalid provider")
	}
}

func (a application) homepageHandler(w http.ResponseWriter, r *http.Request) {
	pageState := PageState{LoggedInSpotify: false, LoggedInYoutube: false}

	spotifyProvider, err := a.getProvider(PROVIDER_SPOTIFY, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if spotifyProvider.IsLoggedIn() {
		pageState.LoggedInSpotify = true
	}

	youtubeProvider, err := a.getProvider(PROVIDER_GOOGLE, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if youtubeProvider.IsLoggedIn() {
		pageState.LoggedInYoutube = true
	}

	tmpl := template.Must(loadHomeTemplate())
	err = tmpl.Execute(w, pageState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadHomeTemplate() (*template.Template, error) {
	name := "./ui/html/home.page.tmpl"
	return template.New(filepath.Base(name)).ParseFiles(name)
}

func (a application) authCallbackHandler(w http.ResponseWriter, r *http.Request) {

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, "provider was not specified", http.StatusInternalServerError)
		return
	}

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokens := oauth2.Token{AccessToken: user.AccessToken, RefreshToken: user.RefreshToken}
	err = a.sessionManager.UpdateTokens(provider, &tokens, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type PageState struct {
	SpotifyUser     string
	LoggedInSpotify bool
	LoggedInYoutube bool
}
