package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/markbates/goth/gothic"
	"github.com/paulombcosta/waltz/provider"
	"github.com/paulombcosta/waltz/provider/spotify"
	"github.com/paulombcosta/waltz/provider/youtube"
	"github.com/paulombcosta/waltz/token"
	"github.com/paulombcosta/waltz/transfer"
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
	Playlists []TransferPlaylist `json:"playlists"`
}

func (t TransferPayload) ToProviderPlaylist() []provider.Playlist {
	providerPlaylist := []provider.Playlist{}
	for _, p := range t.Playlists {
		providerPlaylist = append(providerPlaylist, provider.Playlist{
			ID:   provider.PlaylistID(p.ID),
			Name: p.Name,
		})
	}
	return providerPlaylist
}

type TransferPlaylist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{}

func (a application) transferHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	publisher := transfer.NewWebSocketProgressPublisher(c)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			publisher.Error(err.Error())
			break
		}
		payload, err := parseMessage(message)
		if err != nil {
			publisher.Error(err.Error())
			break
		}
		if len(payload.Playlists) == 0 {
			publisher.Error("failure: no playlists selected")
			break
		}
		origin, err := a.getProvider(PROVIDER_SPOTIFY, r, w)
		if err != nil {
			publisher.Error(err.Error())
			break
		}

		destination, err := a.getProvider(PROVIDER_GOOGLE, r, w)
		if err != nil {
			publisher.Error(err.Error())
			break
		}

		err = transfer.Transfer().
			Playlists(payload.ToProviderPlaylist()).
			From(origin).
			To(destination).
			WithProgressPublisher(publisher).
			Build().Start()

		if err != nil {
			publisher.Error(err.Error())
			break
		}
	}
}

func parseMessage(payload []byte) (*TransferPayload, error) {
	var data TransferPayload
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return nil, err
	}
	return &data, err
}

func (a application) getProvider(name string, r *http.Request, w http.ResponseWriter) (provider.Provider, error) {
	tokenProvider := token.New(name, r, w, a.sessionManager)
	if name == PROVIDER_GOOGLE {
		return youtube.NewApiProvider(tokenProvider), nil
	} else if name == PROVIDER_SPOTIFY {
		return spotify.New(tokenProvider), nil
	} else {
		return nil, fmt.Errorf("invalid provider %s", name)
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
		var content PlaylistsContent
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
		tmpl := template.Must(loadPage("playlist"))
		err = tmpl.Execute(w, pageState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tmpl := template.Must(loadPage("login"))
		err = tmpl.Execute(w, pageState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func loadPage(templateName string) (*template.Template, error) {
	name := fmt.Sprintf("./ui/html/%s.page.tmpl", templateName)
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
