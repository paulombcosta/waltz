package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

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
	store = sessions.NewCookieStore([]byte("1234"))
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistReadPrivate))
	// TODO set a proper state
	state  = "abc123"
	client *spotify.Client
)

type PageState struct {
	SpotifyUser     string
	LoggedInSpotify bool
	LoggedInYoutube bool
}

func main() {
	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:8080/callback/google", "email", "https://www.googleapis.com/auth/youtube"),
	)

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Get("/", http.HandlerFunc(homepageHandler))
	router.Get("/spotify/login", http.HandlerFunc(spotifyLoginHandler))
	// router.Handle("/callback", http.HandlerFunc(completeAuth))
	router.Handle("/auth/{provider}", http.HandlerFunc(startGoogleAuth))
	router.Handle("/callback/google", http.HandlerFunc(googleAuthCallback))
	log.Println("starting server on :8080")
	http.ListenAndServe(":8080", router)
}

func startGoogleAuth(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	usr := session.Values[GOOGLE_USER_TOKEN_SESSION_KEY]
	log.Println("user", usr)
	if usr != nil {
		log.Println("user present, refreshing token")
		provider, err := goth.GetProvider("google")
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get google provider due to: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		user := usr.(goth.User)
		log.Println("user: ", user)
		log.Println("user acccess token", user.AccessToken)
		log.Println("user refresh token", user.RefreshToken)

		log.Println("refresh token:, ", user.RefreshToken)
		updatedToken, err := provider.RefreshToken(user.RefreshToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get google provider due to: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		user.AccessToken = updatedToken.AccessToken
		user.RefreshToken = updatedToken.RefreshToken

		session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = user
		session.Save(r, w)

		source := TokenSource{User: usr.(goth.User)}

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

func googleAuthCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	log.Println("AFTER COMPLETE AUTH")
	log.Println("user acess token, ", user.AccessToken)
	log.Println("user refresh token, ", user.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, _ := store.Get(r, SESSION_NAME)
	session.Values[GOOGLE_USER_TOKEN_SESSION_KEY] = user
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type TokenSource struct {
	User goth.User
}

func (s TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  s.User.AccessToken,
		RefreshToken: s.User.RefreshToken,
	}, nil
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
	pageState := PageState{LoggedInSpotify: false, LoggedInYoutube: false}
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
