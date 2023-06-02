package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/paulombcosta/waltz/provider"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeApiProvider struct {
	tokenProvider provider.TokenProvider
}

const (
	API_ENDPOINT = "http://localhost:8000/"
)

type SearchResponse struct {
	VideoId string `json:"video_id"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}

type playlistResponse struct {
	Title      string `json:"title"`
	PlaylistId string `json:"playlistId"`
}

type fullPlaylistResponse struct {
	Title      string             `json:"title"`
	ID         string             `json:"id"`
	TrackCount uint32             `json:"trackCount"`
	Author     fullPlaylistAuthor `json:"author"`
	Tracks     []playlistTrack    `json:"tracks"`
}

type fullPlaylistAuthor struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type playlistTrack struct {
	ID      string           `json:"videoId"`
	Title   string           `json:"title"`
	Artists []playlistArtist `json:"artists"`
}

type playlistArtist struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type apiAuthentication struct {
	RefreshToken string  `json:"refresh_token"`
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresAt    int64   `json:"expires_at"`
	ExpiresIn    float64 `json:"expires_in"`
}

type createPlaylistPayload struct {
	Title string            `json:"title"`
	Auth  apiAuthentication `json:"auth"`
}

type createPlaylistResponse struct {
	ID string `json:"id"`
}

type createPlaylistItemPayload struct {
	PlaylistId string            `json:"playlist_id"`
	TrackId    string            `json:"track_id"`
	Auth       apiAuthentication `json:"auth"`
}

type createPlaylistItemResponse struct {
	Status string `json:"status"`
}

var ErrorTrackNotFound = errors.New("track not found")

func (y YoutubeApiProvider) Name() string {
	return "YouTubeApi"
}

// maybe move to sessions, looks more like it
func (y YoutubeApiProvider) IsLoggedIn() bool {
	_, err := y.tokenProvider.RefreshToken()
	return err == nil
}

func NewApiProvider(tokenProvider provider.TokenProvider) *YoutubeApiProvider {
	return &YoutubeApiProvider{tokenProvider: tokenProvider}
}

func (y YoutubeApiProvider) createAuthPayload() (*apiAuthentication, error) {
	token, err := y.tokenProvider.GetToken()
	if err != nil {
		return nil, err
	}

	expiresAt := token.Expiry.Unix()
	expiresIn := time.Until(token.Expiry).Seconds()

	return &apiAuthentication{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
		ExpiresIn:    expiresIn,
	}, nil
}

func (y YoutubeApiProvider) CreatePlaylist(name string) (provider.PlaylistID, error) {

	auth, err := y.createAuthPayload()

	if err != nil {
		return "", err
	}

	payload := createPlaylistPayload{
		Title: name,
		Auth:  *auth,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	response, err := http.Post(API_ENDPOINT+"/playlist", "application/json", bytes.NewBuffer(jsonPayload))

	if err != nil {
		return "", err
	}

	responsePayload := createPlaylistResponse{}

	err = json.NewDecoder(response.Body).Decode(&responsePayload)

	if err != nil {
		return "", err
	}

	return provider.PlaylistID(responsePayload.ID), nil
}

func (y YoutubeApiProvider) AddToPlaylist(playlistId string, trackId string) error {
	auth, err := y.createAuthPayload()

	if err != nil {
		return err
	}

	payload := createPlaylistItemPayload{
		PlaylistId: playlistId,
		TrackId:    trackId,
		Auth:       *auth,
	}

	payloadJson, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	res, err := http.Post(API_ENDPOINT+"/track", "application/json", bytes.NewBuffer(payloadJson))

	if err != nil {
		return err
	}

	responsePayload := createPlaylistItemResponse{}

	err = json.NewDecoder(res.Body).Decode(&responsePayload)

	if err != nil {
		return err
	}
	log.Println("insert playlist item status: " + responsePayload.Status)
	return nil
}

func (y YoutubeApiProvider) getYoutubeClient() (*youtube.Service, error) {
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

func (y YoutubeApiProvider) GetPlaylists() ([]provider.Playlist, error) {
	client, err := y.getYoutubeClient()

	if err != nil {
		return nil, err
	}

	channelIdCall := client.Channels.List([]string{"id"}).MaxResults(1).Mine(true)

	res, err := channelIdCall.Do()

	if err != nil {
		return nil, err
	}

	log.Println("channel id = ", res.Items[0].Id)

	playlistRes, err := http.Get(API_ENDPOINT + "/playlists/" + res.Items[0].Id)

	if err != nil {
		return nil, err
	}

	data := []playlistResponse{}

	json.NewDecoder(playlistRes.Body).Decode(&data)

	providerPlaylist := []provider.Playlist{}

	for _, p := range data {
		providerPlaylist = append(providerPlaylist, provider.Playlist{
			Name: p.Title,
			ID:   provider.PlaylistID(p.PlaylistId),
		})
	}

	return providerPlaylist, nil
}

func (y YoutubeApiProvider) GetFullPlaylist(id string) (*provider.FullPlaylist, error) {
	res, err := http.Get(API_ENDPOINT + "/playlist/" + id)
	if err != nil {
		return nil, err
	}

	playlistResponse := fullPlaylistResponse{}

	err = json.NewDecoder(res.Body).Decode(&playlistResponse)

	if err != nil {
		return nil, err
	}

	fullPlaylist := provider.FullPlaylist{
		Playlist: provider.Playlist{
			ID:      provider.PlaylistID(playlistResponse.ID),
			Name:    playlistResponse.Title,
			Creator: playlistResponse.Author.Name,
		},
	}

	tracks := []provider.Track{}

	for _, t := range playlistResponse.Tracks {

		artists := []string{}
		for _, a := range t.Artists {
			artists = append(artists, a.Name)
		}

		track := provider.Track{
			ID:      t.ID,
			Name:    t.Title,
			Artists: artists,
		}

		tracks = append(tracks, track)
	}

	fullPlaylist.Tracks = tracks

	return &fullPlaylist, nil
}

func (y YoutubeApiProvider) FindPlaylistByName(name string) (provider.PlaylistID, error) {
	playlists, err := y.GetPlaylists()
	if err != nil {
		return "", err
	}
	for _, p := range playlists {
		if p.Name == name {
			return p.ID, nil
		}
	}
	return "", nil
}

func (y YoutubeApiProvider) FindTrack(name string) (provider.TrackID, error) {
	res, err := searchTrack(name)
	if err != nil {
		return "", err
	}
	return res, nil
}

func searchTrack(name string) (provider.TrackID, error) {
	res, err := http.Get(API_ENDPOINT + "/track/search" + "?q=" + url.QueryEscape(name))
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return "", ErrorTrackNotFound
		}
		errorResponse := ErrorResponse{}
		json.NewDecoder(res.Body).Decode(&errorResponse)
		return "", errors.New(errorResponse.Detail)
	}
	result := SearchResponse{}
	json.NewDecoder(res.Body).Decode(&result)
	return provider.TrackID(result.VideoId), nil
}
