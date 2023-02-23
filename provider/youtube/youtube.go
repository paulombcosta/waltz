package youtube

import (
	"context"
	"net/http"

	"github.com/paulombcosta/waltz/provider"
	"github.com/paulombcosta/waltz/session"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeProvider struct {
	sessionManager session.SessionManager
}

func New(sessionManager session.SessionManager) YoutubeProvider {
	return YoutubeProvider{sessionManager: sessionManager}
}

// maybe move to sessions, looks more like it
func (y YoutubeProvider) IsLoggedIn(r *http.Request, w http.ResponseWriter) bool {
	tokens, err := y.sessionManager.GetGoogleTokens(r)
	if err != nil {
		return false
	}
	if tokens == nil {
		return false
	}
	_, err = y.sessionManager.RefreshToken("google", r, w)
	if err != nil {
		return false
	}
	return true
}

func (y YoutubeProvider) GetPlaylists(r *http.Request) ([]provider.Playlist, error) {
	tokens, err := y.sessionManager.GetGoogleTokens(r)
	if err != nil {
		return nil, err
	}
	client, err := getYoutubeClient(tokens)
	if err != nil {
		return nil, err
	}

	call := client.Playlists.List([]string{"snippet", "id", "contentDetails"})
	call.Mine(true)
	res, err := call.Do()
	if err != nil {
		return nil, err
	}
	playlists := []provider.Playlist{}
	for _, p := range res.Items {
		playlists = append(playlists, provider.Playlist{Name: p.Snippet.Title})
	}
	return playlists, nil
}

func getYoutubeClient(tokens *oauth2.Token) (*youtube.Service, error) {
	source := TokenSource{Source: *tokens}
	youtubeService, err := youtube.NewService(
		context.Background(), option.WithTokenSource(source))
	if err != nil {
		return nil, err
	}
	return youtubeService, nil
}

type TokenSource struct {
	Source oauth2.Token
}

func (s TokenSource) Token() (*oauth2.Token, error) {
	return &s.Source, nil
}
