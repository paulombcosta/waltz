package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	SPOTIFY_TOKEN_SESSION_KEY     = "spotify-token"
	GOOGLE_USER_TOKEN_SESSION_KEY = "google-user"
	SESSION_NAME                  = "token-session"
)

// extract this to youtube.go eventually
func getYoutubePlaylists(client *youtube.Service) error {
	call := client.Playlists.List([]string{"snippet", "id", "contentDetails"})
	call.Mine(true)
	res, err := call.Do()
	if err != nil {
		return err
	}
	for _, p := range res.Items {
		log.Println("playlist title: ", p.Snippet.Title)
		log.Println("playlist kind: ", p.Kind)
	}
	log.Println("result ", res)
	log.Printf("playlist response = %v", res)
	return nil
}

func getYoutubeClient(r *http.Request, w http.ResponseWriter, store *sessions.CookieStore) (*youtube.Service, error) {
	session, _ := store.Get(r, SESSION_NAME)
	tok := session.Values[GOOGLE_USER_TOKEN_SESSION_KEY]
	if tok != nil {
		newTokens, err := refreshToken("google", r, w, store)
		if err != nil {
			return nil, err
		}
		source := TokenSource{Source: *newTokens}
		youtubeService, err := youtube.NewService(
			context.Background(), option.WithTokenSource(source))
		if err != nil {
			return nil, err
		}
		return youtubeService, nil
	}
	return nil, nil
}

func getSessionTokens(provider string, r *http.Request, store *sessions.CookieStore) (*oauth2.Token, error) {
	session, err := store.Get(r, SESSION_NAME)
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

func updateSession(session *sessions.Session, provider string, tokens *oauth2.Token, r *http.Request, w http.ResponseWriter) error {
	if provider == "spotify" {
		session.Values[SPOTIFY_TOKEN_SESSION_KEY] = tokens
	} else if provider == "google" {
		session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = tokens
	} else {
		return errors.New(fmt.Sprintf("invalid provider %s", provider))
	}
	return session.Save(r, w)
}

func refreshToken(
	providerName string,
	r *http.Request,
	w http.ResponseWriter,
	store *sessions.CookieStore) (*oauth2.Token, error) {

	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		return nil, err
	}
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	existingTokens, err := getSessionTokens(providerName, r, store)
	if err != nil {
		return nil, err
	}
	newTokens, err := provider.RefreshToken(existingTokens.RefreshToken)
	if err != nil {
		return nil, err
	}
	err = updateSession(session, providerName, newTokens, r, w)
	if err != nil {
		return nil, err
	}
	return newTokens, nil
}

func (a application) homepageHandler(w http.ResponseWriter, r *http.Request) {
	pageState := PageState{LoggedInSpotify: false, LoggedInYoutube: false}

	// Extract this to spotify.go
	spotifyClient, err := getSpotifyClient(r, w, a.store)
	if err != nil {
		log.Println("error getting spotify client: ", err.Error())
	}
	if spotifyClient != nil {
		log.Println("spotify client has been initialized")
		user, err := client.CurrentUser(context.Background())
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
	youtubeClient, err := getYoutubeClient(r, w, a.store)
	if err != nil {
		log.Println("error getting yotubue client, ", err.Error())
	}
	if youtubeClient != nil {
		err = getYoutubePlaylists(youtubeClient)
		if err != nil {
			log.Println("error getting youtube playlists", err.Error())
		}
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

func getSpotifyClient(r *http.Request, w http.ResponseWriter, store *sessions.CookieStore) (*spotify.Client, error) {
	session, _ := store.Get(r, SESSION_NAME)
	tok := session.Values[SPOTIFY_TOKEN_SESSION_KEY]
	if tok != nil {
		newTokens, err := refreshToken("spotify", r, w, store)
		if err != nil {
			return nil, err
		}
		client = spotify.New(spotifyauth.New().Client(r.Context(), newTokens))
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
	session, _ := a.store.Get(r, SESSION_NAME)

	tokens := oauth2.Token{AccessToken: user.AccessToken, RefreshToken: user.RefreshToken}
	err = updateSession(session, provider, &tokens, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type TokenSource struct {
	Source oauth2.Token
}

func (s TokenSource) Token() (*oauth2.Token, error) {
	return &s.Source, nil
}

type PageState struct {
	SpotifyUser     string
	LoggedInSpotify bool
	LoggedInYoutube bool
}
