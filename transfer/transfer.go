package transfer

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/paulombcosta/waltz/provider"
	"github.com/paulombcosta/waltz/provider/youtube"
)

const (
	PROGRESS_STARTED_PLAYLSIT = "playlist-start"
	PROGRESS_PLAYLIST_DONE    = "playlist-done"
	PROGRESS_TRACK_DONE       = "track-done"
	PROGRESS_TRANSFER_DONE    = "done"
	PROGRESS_TRANFER_ERROR    = "error"
)

type ProgressMessage struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

type TransferClientBuilder struct {
	origin      provider.Provider
	playlists   []provider.Playlist
	publisher   ProgressPublisher
	destination provider.Provider
}

func Transfer() TransferClientBuilder {
	return TransferClientBuilder{}
}

func (t TransferClientBuilder) Playlists(playlists []provider.Playlist) TransferClientBuilder {
	t.playlists = playlists
	return t
}

func (t TransferClientBuilder) From(origin provider.Provider) TransferClientBuilder {
	t.origin = origin
	return t
}

func (t TransferClientBuilder) To(destination provider.Provider) TransferClientBuilder {
	t.destination = destination
	return t
}

func (t TransferClientBuilder) WithProgressPublisher(p ProgressPublisher) TransferClientBuilder {
	t.publisher = p
	return t
}

// TODO validate fields here
func (t TransferClientBuilder) Build() TransferClient {
	return TransferClient(t)
}

func NewWebSocketProgressPublisher(conn *websocket.Conn) WebSocketProgressPublisher {
	return WebSocketProgressPublisher{Conn: conn}
}

type WebSocketProgressPublisher struct {
	Conn *websocket.Conn
}

func (publisher WebSocketProgressPublisher) Publish(progressType string, body string) error {
	payload := ProgressMessage{
		Type: progressType,
		Body: body,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return publisher.Conn.WriteMessage(websocket.TextMessage, data)
}

func (publisher WebSocketProgressPublisher) Error(body string) {
	_ = publisher.Publish(PROGRESS_TRANFER_ERROR, body)
}

type ProgressPublisher interface {
	Publish(progressType string, body string) error
}

type TransferClient struct {
	origin      provider.Provider
	playlists   []provider.Playlist
	publisher   ProgressPublisher
	destination provider.Provider
}

func (t TransferClient) publish(typeOf string, content string) {
	_ = t.publisher.Publish(typeOf, content)
}

func (t TransferClient) Start() error {

	if t.playlists == nil {
		return errors.New("cannot import: list is null")
	}
	if len(t.playlists) == 0 {
		return errors.New("cannot import: list is empty")
	}

	for _, playlist := range t.playlists {
		destinationPlaylistId, err := getOrCreatePlaylist(t.destination, playlist)
		if err != nil {
			return err
		}

		t.publish(PROGRESS_STARTED_PLAYLSIT, playlist.Name)
		fullPlaylist, err := t.origin.GetFullPlaylist(string(playlist.ID))
		if err != nil {
			return err
		}
		err = t.addTracksToPlaylist(t.destination, destinationPlaylistId, fullPlaylist.Tracks)
		if err != nil {
			return err
		}
		t.publish(PROGRESS_PLAYLIST_DONE, "")
	}
	t.publish(PROGRESS_TRANSFER_DONE, "")

	return nil
}

func (client TransferClient) addTracksToPlaylist(provider provider.Provider, playlistId string, tracks []provider.Track) error {
	currentPlaylist, err := provider.GetFullPlaylist(playlistId)
	if err != nil {
		return err
	}
	existingTracks := currentPlaylist.Tracks

	for _, t := range tracks {

		trackId, err := provider.FindTrack(t.FullName())
		if err != nil {
			if errors.Is(err, youtube.ErrorTrackNotFound) {
				log.Printf("track %s not found, skipping it", t.FullName())
				client.publish(PROGRESS_TRACK_DONE, "")
				continue
			}
			log.Println(err)
			return err
		}

		log.Println("found track with id ", trackId)

		if trackId == "" {
			client.publish(PROGRESS_TRACK_DONE, "")
			continue
		}

		// See if playlist already has an item with the videoID
		isDuplicate := false
		for _, t := range existingTracks {
			if string(trackId) == t.ID {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			client.publish(PROGRESS_TRACK_DONE, "")
			continue
		}

		err = provider.AddToPlaylist(playlistId, string(trackId))
		if err != nil {
			client.publish(PROGRESS_TRACK_DONE, "")
			return err
		}

		client.publish(PROGRESS_TRACK_DONE, "")
	}
	return nil
}

func getOrCreatePlaylist(destination provider.Provider, playlist provider.Playlist) (string, error) {
	destinationPlaylist := ""
	id, err := destination.FindPlaylistByName(string(playlist.Name))
	if err != nil {
		return "", err
	}
	if id == "" {
		id, err = destination.CreatePlaylist(playlist.Name)
		if err != nil {
			return "", err
		}
		destinationPlaylist = string(id)
	} else {
		destinationPlaylist = string(id)
	}
	return destinationPlaylist, nil
}
