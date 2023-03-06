package youtube

import (
	"context"

	"github.com/paulombcosta/waltz/provider"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeProvider struct {
	tokenProvider provider.TokenProvider
}

func New(tokenProvider provider.TokenProvider) *YoutubeProvider {
	return &YoutubeProvider{tokenProvider: tokenProvider}
}

// maybe move to sessions, looks more like it
func (y YoutubeProvider) IsLoggedIn() bool {
	_, err := y.tokenProvider.RefreshToken()
	return err == nil
}

func (y YoutubeProvider) FindTrack(name string) (*provider.TrackID, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil, err
	}
	searchResponse, err := client.Search.List([]string{"id"}).Type("video").MaxResults(1).Q(name).Do()
	if err != nil {
		return nil, err
	}
	return (*provider.TrackID)(&searchResponse.Items[0].Id.VideoId), nil
}

func (y YoutubeProvider) FindPlaylist(name string) (*provider.PlaylistID, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil, err
	}
	searchResponse, err := client.Search.List([]string{"id"}).Type("playlist").MaxResults(1).Q(name).Do()
	if err != nil {
		return nil, err
	}
	return (*provider.PlaylistID)(&searchResponse.Items[0].Id.PlaylistId), nil
}

func (y YoutubeProvider) CreatePlaylist(name string) (*provider.PlaylistID, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil, err
	}
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       name,
			Description: "Playlist imported by Waltz",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	}

	playlist, err = client.Playlists.Insert([]string{"snippet", "status"}, playlist).Do()
	if err != nil {
		return nil, err
	}
	return (*provider.PlaylistID)(&playlist.Id), nil
}

func (y YoutubeProvider) GetPlaylists() ([]provider.Playlist, error) {
	client, err := y.getYoutubeClient()
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

func (y YoutubeProvider) getYoutubeClient() (*youtube.Service, error) {
	tokens, err := y.tokenProvider.GetToken()
	if err != nil {
		return nil, err
	}
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
