package transfer

import (
	"log"
	"testing"

	"github.com/paulombcosta/waltz/provider"
)

func TestShouldReturnErrorIfPlaylistsAreNil(t *testing.T) {
	origin := provider.NewMockProvider(t)
	destination := provider.NewMockProvider(t)
	err := Transfer(origin, nil).To(destination)
	expectedMsg := "cannot import: list is null"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldReturnErrorIfPlaylistsAreEmpty(t *testing.T) {
	origin := provider.NewMockProvider(t)
	destination := provider.NewMockProvider(t)
	err := Transfer(origin, []provider.Playlist{}).To(destination)
	expectedMsg := "cannot import: list is empty"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldUseExistingPlaylistOnDestination(t *testing.T) {
	destination := provider.NewMockProvider(t)

	destinationPlaylist := provider.Playlist{ID: "123"}

	// destination.EXPECT().FindPlaylistByName("id").Return(&destinationPlaylist, nil).Once()

	id, _ := getOrCreatePlaylist(destination, destinationPlaylist)

	if id != "123" {
		log.Fatalf("expected id to be 123 but it is %s", id)
	}
}
