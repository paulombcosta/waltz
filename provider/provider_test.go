package provider

import (
	"testing"
)

func TestTrackFullName(t *testing.T) {
	track := Track{
		Name:    "Song",
		Artists: []string{"Paulo", "Other"},
	}
	expected := "Paulo, Other - Song"
	actual := track.FullName()
	if actual != expected {
		t.Fatalf("expected %s but got %s", expected, actual)
	}
}
