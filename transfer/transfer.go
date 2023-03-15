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

	log.Printf("starting transfer from %s to %s", t.Origin.Name(), destination.Name())

	if t.playlists == nil {
		return errors.New("cannot import: list is null")
	}
	if len(t.playlists) == 0 {
		return errors.New("cannot import: list is empty")
	}

	log.Println("fetching playlists")
	for _, playlist := range t.playlists {
		// Find if playlist already exists on destination
		destinationPlaylistId, err := getOrCreatePlaylist(destination, playlist)
		if err != nil {
			return err
		}

		log.Printf("fetching tracks on %s for playlist %s", t.Origin.Name(), playlist.Name)
		fullPlaylist, err := t.Origin.GetFullPlaylist(string(playlist.ID))
		if err != nil {
			return err
		}
		err = destination.AddToPlaylist(destinationPlaylistId, fullPlaylist.Tracks)
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
