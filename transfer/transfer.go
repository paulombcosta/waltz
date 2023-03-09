package transfer

import (
	"errors"
	"log"

	"github.com/paulombcosta/waltz/provider"
)

type TransferClient struct {
	Origin    provider.Provider
	playlists []provider.Playlist
}

func Transfer(provider provider.Provider, playlists []provider.Playlist) TransferClient {
	return TransferClient{Origin: provider, playlists: playlists}
}

func (t TransferClient) To(destination provider.Provider) error {
	if t.playlists == nil {
		return errors.New("cannot import: list is null")
	}
	if len(t.playlists) == 0 {
		return errors.New("cannot import: list is empty")
	}

	for _, playlist := range t.playlists {
		// Find if playlist already exists on destination
		_, err := getOrCreatePlaylist(destination, playlist)
		if err != nil {
			return err
		}
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
		log.Println("playlist not found, creating a new one")
		id, err = destination.CreatePlaylist(playlist.Name)
		if err != nil {
			return "", err
		}
		log.Println("playlist created with id ", id)
		destinationPlaylist = string(id)
	} else {
		destinationPlaylist = string(id)
	}
	return destinationPlaylist, nil
}
