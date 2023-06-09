package youtube

import (
	"context"
	"fmt"

	"github.com/paulombcosta/waltz/provider"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeV3Provider struct {
	tokenProvider provider.TokenProvider
	playlists     []*youtube.Playlist
}

func (y YoutubeV3Provider) getPlaylists() ([]*youtube.Playlist, error) {
	if y.playlists != nil {
		return y.playlists, nil
	}
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil, err
	}
	response, err := client.Playlists.List([]string{"snippet", "id"}).Mine(true).Do()
	if err != nil {
		return nil, err
	}
	y.playlists = response.Items
	return y.playlists, nil
}

func NewV3Provider(tokenProvider provider.TokenProvider) *YoutubeV3Provider {
	return &YoutubeV3Provider{tokenProvider: tokenProvider}
}

// maybe move to sessions, looks more like it
func (y YoutubeV3Provider) IsLoggedIn() bool {
	_, err := y.tokenProvider.RefreshToken()
	return err == nil
}

func (y YoutubeV3Provider) FindTrack(name string) (provider.TrackID, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return "", err
	}
	searchResponse, err := client.Search.List([]string{"id"}).Type("video").MaxResults(1).Q(name).Do()
	if err != nil {
		return "", err
	}
	if len(searchResponse.Items) == 0 {
		return "", nil
	}
	return provider.TrackID(searchResponse.Items[0].Id.VideoId), nil
}

func (y YoutubeV3Provider) FindPlaylistByName(name string) (provider.PlaylistID, error) {
	playlists, err := y.getPlaylists()
	if err != nil {
		return "", err
	}
	for _, p := range playlists {
		if p.Snippet.Title == name {
			return provider.PlaylistID(p.Id), nil
		}
	}
	return "", nil
}

func (y YoutubeV3Provider) CreatePlaylist(name string) (provider.PlaylistID, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return "", err
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
		return "", err
	}
	return provider.PlaylistID(playlist.Id), nil
}

func (y YoutubeV3Provider) GetPlaylists() ([]provider.Playlist, error) {
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

func (y YoutubeV3Provider) Name() string {
	return "YouTubeV3"
}

func (y YoutubeV3Provider) GetFullPlaylist(id string) (*provider.FullPlaylist, error) {
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil, err
	}
	playlist := &provider.FullPlaylist{
		Playlist: provider.Playlist{ID: provider.PlaylistID(id)},
	}
	tracks := []provider.Track{}
	nextPageToken := ""
	for {
		playlistItemListCall := client.PlaylistItems.List([]string{"contentDetails"}).
			PlaylistId(id).
			MaxResults(50).
			PageToken(nextPageToken)

		playlistItemListResponse, err := playlistItemListCall.Do()
		if err != nil {
			return nil, fmt.Errorf("error retrieving playlist items: %v", err)
		}

		for _, item := range playlistItemListResponse.Items {
			tracks = append(tracks, provider.Track{
				ID: item.ContentDetails.VideoId,
			})
		}
		nextPageToken = playlistItemListResponse.NextPageToken

		if nextPageToken == "" {
			break
		}
	}
	playlist.Tracks = tracks
	return playlist, nil
}

func (y YoutubeV3Provider) AddToPlaylist(playlistId string, trackId string) error {
	client, err := y.getYoutubeClient()
	if err != nil {
		return nil
	}

	item := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistId,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: string(trackId),
			},
		},
	}
	_, err = client.PlaylistItems.Insert([]string{"snippet"}, item).Do()
	if err != nil {
		return err
	}
	return nil
}

func (y YoutubeV3Provider) getYoutubeClient() (*youtube.Service, error) {
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

func (y YoutubeV3Provider) Debug() {
	panic("not used")
}
