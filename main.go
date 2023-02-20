package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/paulombcosta/waltz/spotifyauth"

	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	ch   = make(chan *spotify.Client)
	// TODO set a proper state
	state = "abc123"
)

type PageState struct {
	LoggedInSpotify bool
	LoggedInGoogle  bool
}

func main() {
	router := chi.NewRouter()
	router.Get("/", http.HandlerFunc(homepageHandler))
	router.Get("/spotify/login", http.HandlerFunc(spotifyLoginHandler))
	log.Println("starting server on :8080")
	http.ListenAndServe(":8080", router)
}

func loadHomeTemplate() (*template.Template, error) {
	name := "./ui/html/home.page.tmpl"
	return template.New(filepath.Base(name)).ParseFiles(name)
}

func spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)
	// r.Header.Set("HX-Redirect", url)
	// w.WriteHeader(http.StatusOK)
	w.Header().Set("HX-Redirect", url)
	// fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// // wait for auth to complete
	// client := <-ch

	// // use the client to make calls that require authorization
	// _, err := client.CurrentUser()
	// if err != nil {
	// 	http.Error(w, "Couldn't get user", http.StatusForbidden)
	// 	log.Fatal(err)
	// }
	// r.Header.Set("HX-Refresh", "true")
	// tmpl := template.Must(loadHomeTemplate())
	// tmpl.Execute(w, PageState{LoggedInSpotify: true, LoggedInGoogle: false})
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	tok, _ := auth.Token(r.Context(), state, r)
	pageState := PageState{LoggedInSpotify: false, LoggedInGoogle: false}
	if tok != nil {
		pageState.LoggedInSpotify = true
	}
	tmpl := template.Must(loadHomeTemplate())
	err := tmpl.Execute(w, pageState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setupHandlers() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
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
	http.Redirect(w, r.WithContext(r.Context()), "/", http.StatusSeeOther)
	// use the token to get an authenticated client
	// client := spotify.NewClient(auth.Client(r.Context(), tok))
	// fmt.Fprintf(w, "Login Completed!")
}
