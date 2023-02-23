package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/markbates/goth/gothic"
	"github.com/paulombcosta/waltz/session"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func (a application) homepageHandler(w http.ResponseWriter, r *http.Request) {
	pageState := PageState{LoggedInSpotify: false, LoggedInYoutube: false}

	// Extract this to spotify.go
	spotifyClient, err := getSpotifyClient(r, w, a.sessionManager)
	if err != nil {
		log.Println("error getting spotify client: ", err.Error())
	}
	if spotifyClient != nil {
		log.Println("spotify client has been initialized")
		user, err := spotifyClient.CurrentUser(context.Background())
		if err != nil {
			log.Println("error getting current user: ", err.Error())
		} else {
			pageState.LoggedInSpotify = true
			pageState.SpotifyUser = user.ID
		}
	} else {
		log.Println("token not found, user not logged in")
	}

	// Extract this to youtube.go
	if a.youtubeProvider.IsLoggedIn(r, w) {
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

func getSpotifyClient(r *http.Request, w http.ResponseWriter, sessionManager session.SessionManager) (*spotify.Client, error) {
	tok, err := sessionManager.GetSpotifyTokens(r)
	if err != nil {
		return nil, err
	}
	if tok != nil {
		newTokens, err := sessionManager.RefreshToken("spotify", r, w)
		if err != nil {
			return nil, err
		}
		client := spotify.New(spotifyauth.New().Client(r.Context(), newTokens))
		return client, nil
	} else {
		return nil, nil
	}
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
