package transfer

import (
	"errors"
	"log"
	"testing"

	"github.com/paulombcosta/waltz/provider"
)

func getMockProvider(t *testing.T) *provider.MockProvider {
	p := provider.NewMockProvider(t)
	p.EXPECT().Name().Return("Test")
	return p
}

func TestShouldReturnErrorIfPlaylistsAreNil(t *testing.T) {
	origin := getMockProvider(t)
	destination := getMockProvider(t)
	err := Transfer(origin, nil).To(destination)
	expectedMsg := "cannot import: list is null"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldReturnErrorIfPlaylistsAreEmpty(t *testing.T) {
	origin := getMockProvider(t)
	destination := getMockProvider(t)
	err := Transfer(origin, []provider.Playlist{}).To(destination)
	expectedMsg := "cannot import: list is empty"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldUseExistingPlaylistOnDestination(t *testing.T) {
	destination := provider.NewMockProvider(t)

	destinationPlaylist := provider.Playlist{ID: "123", Name: "name"}

	destination.EXPECT().FindPlaylistByName("name").Return(provider.PlaylistID("123"), nil).Once()

	id, _ := getOrCreatePlaylist(destination, destinationPlaylist)

	if id != "123" {
		log.Fatalf("expected id to be 123 but it is %s", id)
	}
}

func TestShouldCreatePlaylistWhenOriginDoesntExist(t *testing.T) {
	destination := provider.NewMockProvider(t)

	destinationPlaylist := provider.Playlist{ID: "123", Name: "name"}

	destination.EXPECT().FindPlaylistByName("name").Return("", nil).Once()
	destination.EXPECT().CreatePlaylist("name").Return("123", nil).Once()

	id, _ := getOrCreatePlaylist(destination, destinationPlaylist)

	if id != "123" {
		log.Fatalf("expected id to be 123 but it is %s", id)
	}
}

func TestShouldUseOriginPlaylistIDWhenFetchingFullPlaylist(t *testing.T) {
	origin := getMockProvider(t)
	destination := getMockProvider(t)

	originPlaylistID := "origin-ID"
	destinationPlaylistID := "destination-ID"

	playlists := []provider.Playlist{
		{
			ID:   provider.PlaylistID(originPlaylistID),
			Name: "playlist",
		},
	}

	destination.EXPECT().FindPlaylistByName("playlist").Return(provider.PlaylistID(destinationPlaylistID), nil).Once()
	origin.EXPECT().GetFullPlaylist(originPlaylistID).Return(nil, errors.New("stop")).Once()

	Transfer(origin, playlists).To(destination)
}
