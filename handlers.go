package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
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

type PlaylistsContent struct {
	Playlists []provider.Playlist
	Err       string
}

type PageState struct {
	LoggedInSpotify  bool
	LoggedInYoutube  bool
	PlaylistsContent PlaylistsContent
}

type TransferPayload struct {
	Playlists []string `json:"playlists"`
}

func (a application) transferHandler(w http.ResponseWriter, r *http.Request) {
	var payload TransferPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("received payload", payload)
}

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
	pageState := PageState{
		LoggedInSpotify:  false,
		LoggedInYoutube:  false,
		PlaylistsContent: PlaylistsContent{},
	}

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

	if pageState.LoggedInSpotify && pageState.LoggedInYoutube {
		playlists, err := spotifyProvider.GetPlaylists()
		content := PlaylistsContent{}
		if err != nil {
			content = PlaylistsContent{
				Playlists: []provider.Playlist{},
				Err:       err.Error(),
			}
		} else {
			content = PlaylistsContent{
				Playlists: playlists,
				Err:       "",
			}
		}
		pageState.PlaylistsContent = content
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
	ts, err := template.New(filepath.Base(name)).ParseFiles(name)
	if err != nil {
		return nil, err
	}
	ts, err = ts.ParseGlob(filepath.Join("./ui/html/", "*.layout.tmpl"))
	if err != nil {
		return nil, err
	}
	return ts, nil
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
