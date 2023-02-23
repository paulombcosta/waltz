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
	spotifyProvider "github.com/markbates/goth/providers/spotify"

	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

var (
	// TODO get a proper session key
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
			"http://localhost:8080/auth/callback?provider=google", "email", "https://www.googleapis.com/auth/youtube"),
		spotifyProvider.New(
			os.Getenv("SPOTIFY_ID"),
			os.Getenv("SPOTIFY_SECRET"),
			"http://localhost:8080/auth/callback?provider=spotify",
			"user-read-private", "playlist-read-private"),
	)

	router := chi.NewRouter()

	app := application{store: store}

	router.Get("/", http.HandlerFunc(app.homepageHandler))
	router.Get("/auth", gothic.BeginAuthHandler)
	router.Handle("/auth/callback", http.HandlerFunc(app.authCallbackHandler))
	log.Println("starting server on :8080")
	http.ListenAndServe(":8080", router)
}

func init() {
	gob.Register(&oauth2.Token{})
}
