package transfer

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/paulombcosta/waltz/provider"
)

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

func (t TransferClientBuilder) WithProgressObserver(p ProgressPublisher) TransferClientBuilder {
	t.publisher = p
	return t
}

// TODO validate fields here
func (t TransferClientBuilder) Build() TransferClient {
	return TransferClient(t)
}

type WebSocketProgressPublisher struct {
	Conn *websocket.Conn
}

func (publisher WebSocketProgressPublisher) Publish(progressType string) error {
	return publisher.Conn.WriteMessage(websocket.TextMessage, []byte(progressType))
}

type ProgressPublisher interface {
	Publish(progressType string) error
}

type TransferClient struct {
	origin      provider.Provider
	playlists   []provider.Playlist
	publisher   ProgressPublisher
	destination provider.Provider
}

// func Transfer(provider provider.Provider, playlists []provider.Playlist) TransferClient {
// 	return TransferClient{Origin: provider, playlists: playlists}
// }

func (t TransferClient) Start() error {

	log.Printf("starting transfer from %s to %s", t.origin.Name(), t.destination.Name())

	if t.playlists == nil {
		return errors.New("cannot import: list is null")
	}
	if len(t.playlists) == 0 {
		return errors.New("cannot import: list is empty")
	}

	log.Println("fetching playlists")
	for _, playlist := range t.playlists {
		destinationPlaylistId, err := getOrCreatePlaylist(t.destination, playlist)
		if err != nil {
			return err
		}

		log.Printf("fetching tracks on %s for playlist %s", t.origin.Name(), playlist.Name)
		fullPlaylist, err := t.origin.GetFullPlaylist(string(playlist.ID))
		if err != nil {
			return err
		}
		err = t.destination.AddToPlaylist(destinationPlaylistId, fullPlaylist.Tracks)
		if err != nil {
			return err
		}
		log.Printf("finished importing for %s\n", playlist.Name)
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
		log.Printf("playlist %s not found, creating a new one", playlist.Name)
		id, err = destination.CreatePlaylist(playlist.Name)
		if err != nil {
			return "", err
		}
		log.Println("playlist created with id ", id)
		destinationPlaylist = string(id)
	} else {
		log.Printf("playlist %s already exists", playlist.Name)
		destinationPlaylist = string(id)
	}
	return destinationPlaylist, nil
}
