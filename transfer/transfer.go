package transfer

import (
	"errors"

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
	return nil
}
