package spotify

import (
	"testing"

	"github.com/benzjeremy/spotify-screensaver/store"
)

func TestControllerStateAndColors(t *testing.T) {
	secStore, err := store.NewSecureStore()
	if err != nil {
		t.Fatalf("Failed to create secure store: %v", err)
	}

	ctrl := NewController(secStore)
	state := ctrl.GetPlaybackState()

	if state.Title == "" {
		t.Fatalf("Title should not be empty")
	}
	if state.PrimaryColor == "" {
		t.Fatalf("PrimaryColor should not be empty")
	}
	if state.SecondaryColor == "" {
		t.Fatalf("SecondaryColor should not be empty")
	}
}
