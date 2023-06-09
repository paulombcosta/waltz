package transfer

import (
	"errors"
	"testing"

	"github.com/paulombcosta/waltz/provider"
)

type NoOpPublisher struct{}

func (p NoOpPublisher) Publish(progressType string, body string) error {
	return nil
}

func getMockProvider(t *testing.T) *provider.MockProvider {
	p := provider.NewMockProvider(t)
	return p
}

func TestShouldReturnErrorIfPlaylistsAreNil(t *testing.T) {
	err := Transfer().
		From(getMockProvider(t)).
		To(getMockProvider(t)).
		Build().
		Start()
	expectedMsg := "cannot import: list is null"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldReturnErrorIfPlaylistsAreEmpty(t *testing.T) {
	err := Transfer().
		From(getMockProvider(t)).
		To(getMockProvider(t)).
		Playlists([]provider.Playlist{}).
		Build().
		Start()
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
		t.Fatalf("expected id to be 123 but it is %s", id)
	}
}

func TestShouldCreatePlaylistWhenOriginDoesntExist(t *testing.T) {
	destination := provider.NewMockProvider(t)

	destinationPlaylist := provider.Playlist{ID: "123", Name: "name"}

	destination.EXPECT().FindPlaylistByName("name").Return("", nil).Once()
	destination.EXPECT().CreatePlaylist("name").Return("123", nil).Once()

	id, _ := getOrCreatePlaylist(destination, destinationPlaylist)

	if id != "123" {
		t.Fatalf("expected id to be 123 but it is %s", id)
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

	_ = Transfer().
		From(origin).
		To(destination).
		Playlists(playlists).
		WithProgressPublisher(NoOpPublisher{}).
		Build().
		Start()
}
