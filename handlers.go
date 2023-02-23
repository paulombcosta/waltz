package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type AuthData struct {
}

func (a application) spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)
	w.Header().Set("HX-Redirect", url)
}

func (a application) homepageHandler(w http.ResponseWriter, r *http.Request) {
	pageState := PageState{LoggedInSpotify: false, LoggedInYoutube: false}
	spotifyClient := getSpotifyClient(r, a.store)
	if spotifyClient != nil {
		log.Println("spotify client has been initialized")
		user, err := client.CurrentUser(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pageState.LoggedInSpotify = true
		pageState.SpotifyUser = user.ID
	} else {
		log.Println("token not found, user not logged in")
	}
	tmpl := template.Must(loadHomeTemplate())
	err := tmpl.Execute(w, pageState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadHomeTemplate() (*template.Template, error) {
	name := "./ui/html/home.page.tmpl"
	return template.New(filepath.Base(name)).ParseFiles(name)
}

func getSpotifyClient(r *http.Request, store *sessions.CookieStore) *spotify.Client {
	if client != nil {
		return client
	} else {
		session, _ := store.Get(r, SESSION_NAME)
		tok := session.Values[SPOTIFY_TOKEN_SESSION_KEY]
		if tok != nil {
			client = spotify.New(auth.Client(r.Context(), tok.(*oauth2.Token)))
			return client
		} else {
			return nil
		}
	}
}

func (a application) spotifyAuthCallback(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	log.Println("token:", tok)
	session, _ := a.store.Get(r, SESSION_NAME)
	// only store the bare-minimum necessary
	session.Values[SPOTIFY_TOKEN_SESSION_KEY] = oauth2.Token{AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken}
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a application) startGoogleAuth(w http.ResponseWriter, r *http.Request) {
	session, _ := a.store.Get(r, SESSION_NAME)
	tokens := session.Values[GOOGLE_USER_TOKEN_SESSION_KEY]
	log.Println("tokens", tokens)
	if tokens != nil {
		log.Println("user present, refreshing token")
		provider, err := goth.GetProvider("google")
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get google provider due to: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		t := tokens.(oauth2.Token)

		log.Println("user acccess token", t.AccessToken)
		log.Println("user refresh token", t.RefreshToken)

		updatedToken, err := provider.RefreshToken(t.RefreshToken)

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get google provider due to: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = oauth2.Token{AccessToken: updatedToken.AccessToken, RefreshToken: updatedToken.RefreshToken}
		session.Save(r, w)

		source := TokenSource{Source: *updatedToken}

		youtubeService, err := youtube.NewService(context.Background(), option.WithTokenSource(source))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		call := youtubeService.Playlists.List([]string{"snippet", "id", "contentDetails"})
		call.Mine(true)
		res, err := call.Do()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, p := range res.Items {
			log.Println("playlist title: ", p.Snippet.Title)
			log.Println("playlist kind: ", p.Kind)
		}
		log.Println("result ", res)
		log.Printf("playlist response = %v", res)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (a application) googleAuthCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	log.Println("AFTER COMPLETE AUTH")
	log.Println("user acess token, ", user.AccessToken)
	log.Println("user refresh token, ", user.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, _ := a.store.Get(r, SESSION_NAME)
	session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = oauth2.Token{AccessToken: user.AccessToken, RefreshToken: user.RefreshToken}
	session.Save(r, w)
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
