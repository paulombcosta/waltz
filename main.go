package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	SPOTIFY_TOKEN_SESSION_KEY     = "spotify-token"
	GOOGLE_USER_TOKEN_SESSION_KEY = "google-user"
	redirectURI                   = "http://localhost:8080/callback"
	SESSION_NAME                  = "token-session"
)

var (
	// TODO get a proper session key
	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistReadPrivate))
	// TODO set a proper state
	state  = "abc123"
	client *spotify.Client
)

type application struct {
	store *sessions.CookieStore
}

func main() {
	store := sessions.NewCookieStore([]byte("1234"))
	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:8080/callback/google", "email", "https://www.googleapis.com/auth/youtube"),
	)

	router := chi.NewRouter()

	app := application{store: store}

	router.Get("/", http.HandlerFunc(app.homepageHandler))
	router.Get("/spotify/login", http.HandlerFunc(app.spotifyLoginHandler))
	router.Handle("/callback", http.HandlerFunc(app.spotifyAuthCallback))
	router.Handle("/auth/{provider}", http.HandlerFunc(app.startGoogleAuth))
	router.Handle("/callback/google", http.HandlerFunc(app.googleAuthCallback))
	log.Println("starting server on :8080")
	http.ListenAndServe(":8080", router)
}

func init() {
	gob.Register(&oauth2.Token{})
}
