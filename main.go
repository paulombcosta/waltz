package main

import (
	"context"
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	// "github.com/paulombcosta/waltz/spotifyauth"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	SPOTIFY_TOKEN_SESSION_KEY = "spotify-token"
	redirectURI               = "http://localhost:8080/callback"
	SESSION_NAME              = "token-session"
)

var (
	// TODO get a proper session key
	store = sessions.NewCookieStore([]byte("1234"))
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	// TODO set a proper state
	state  = "abc123"
	client *spotify.Client
)

type PageState struct {
	SpotifyUser     string
	LoggedInSpotify bool
	LoggedInGoogle  bool
}

func main() {
	router := chi.NewRouter()
	router.Get("/", http.HandlerFunc(homepageHandler))
	router.Get("/spotify/login", http.HandlerFunc(spotifyLoginHandler))
	router.Handle("/callback", http.HandlerFunc(completeAuth))
	log.Println("starting server on :8080")
	http.ListenAndServe(":8080", router)
}

func loadHomeTemplate() (*template.Template, error) {
	name := "./ui/html/home.page.tmpl"
	return template.New(filepath.Base(name)).ParseFiles(name)
}

func spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)
	w.Header().Set("HX-Redirect", url)
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	pageState := PageState{LoggedInSpotify: false, LoggedInGoogle: false}
	spotifyClient := getSpotifyClient(r)
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

func init() {
	gob.Register(&oauth2.Token{})
}

func getSpotifyClient(r *http.Request) *spotify.Client {
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

func completeAuth(w http.ResponseWriter, r *http.Request) {
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
	session, _ := store.Get(r, SESSION_NAME)
	session.Values[SPOTIFY_TOKEN_SESSION_KEY] = tok
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
