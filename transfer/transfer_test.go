package transfer

import (
	"testing"

	"github.com/paulombcosta/waltz/provider"
	"github.com/paulombcosta/waltz/provider/mocks"
)

func TestShouldReturnErrorIfPlaylistsAreNil(t *testing.T) {
	origin := &mocks.Provider{}
	destination := mocks.Provider{}
	err := Transfer(origin, nil).To(&destination)
	expectedMsg := "cannot import: list is null"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}

func TestShouldReturnErrorIfPlaylistsAreEmpty(t *testing.T) {
	origin := &mocks.Provider{}
	destination := mocks.Provider{}
	err := Transfer(origin, []provider.Playlist{}).To(&destination)
	expectedMsg := "cannot import: list is empty"
	actual := err.Error()
	if actual != expectedMsg {
		t.Fatalf("expected error msg to be %s but it is %s", expectedMsg, actual)
	}
}
