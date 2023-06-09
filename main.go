package main

import (
	"encoding/gob"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	spotifyProvider "github.com/markbates/goth/providers/spotify"

	"github.com/paulombcosta/waltz/log"
	"github.com/paulombcosta/waltz/session"
	"golang.org/x/oauth2"
)

type application struct {
	sessionManager session.SessionManager
	TransferData   *transferData
}

func main() {
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

	fileServer := http.FileServer(http.Dir("./ui/static"))

	sessionManager := session.New()

	err := log.Init()

	if err != nil {
		panic("error initializing logger")
	}

	app := application{
		sessionManager: sessionManager,
		TransferData:   &transferData{},
	}

	router.Get("/", http.HandlerFunc(app.homepageHandler))
	router.Get("/auth", gothic.BeginAuthHandler)
	router.Handle("/auth/callback", http.HandlerFunc(app.authCallbackHandler))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))
	router.HandleFunc("/transferSocket", http.HandlerFunc(app.transferSocket))
	router.Post("/transfer", http.HandlerFunc(app.transferHandler))
	log.Logger.Info("starting server on :8080")
	log.Logger.Panic(http.ListenAndServe(":8080", router))
}

func init() {
	gob.Register(&oauth2.Token{})
}
