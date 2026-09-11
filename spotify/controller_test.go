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

func TestDetectAdvertisement(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		artist     string
		album      string
		trackID    string
		wantAd     bool
		wantTitle  string
	}{
		{
			name:      "Normal track",
			title:     "Starlight",
			artist:    "Muse",
			album:     "Black Holes and Revelations",
			trackID:   "spotify:track:4edh85v",
			wantAd:    false,
			wantTitle: "",
		},
		{
			name:      "Spotify ad track ID",
			title:     "Advertisement",
			artist:    "Spotify",
			album:     "",
			trackID:   "spotify:ad:0000001",
			wantAd:    true,
			wantTitle: "Spotify Werbung",
		},
		{
			name:      "English Advertisement title",
			title:     "Advertisement",
			artist:    "Spotify",
			album:     "Spotify",
			trackID:   "",
			wantAd:    true,
			wantTitle: "Spotify Werbung",
		},
		{
			name:      "German Werbung title",
			title:     "Werbung",
			artist:    "",
			album:     "",
			trackID:   "",
			wantAd:    true,
			wantTitle: "Spotify Werbung",
		},
		{
			name:      "Album is Advertisement with commercial title",
			title:     "Premium Upgrade Angebot",
			artist:    "Spotify",
			album:     "Advertisement",
			trackID:   "",
			wantAd:    true,
			wantTitle: "Premium Upgrade Angebot",
		},
		{
			name:      "TrackID with :ad: embedded",
			title:     "Audio Interstitial",
			artist:    "Advertiser",
			album:     "Commercials",
			trackID:   "urn:spotify:ad:xyz",
			wantAd:    true,
			wantTitle: "Audio Interstitial",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotAd, gotTitle := DetectAdvertisement(tc.title, tc.artist, tc.album, tc.trackID)
			if gotAd != tc.wantAd {
				t.Errorf("DetectAdvertisement() gotAd = %v, want %v", gotAd, tc.wantAd)
			}
			if tc.wantAd && gotTitle != tc.wantTitle {
				t.Errorf("DetectAdvertisement() gotTitle = %q, want %q", gotTitle, tc.wantTitle)
			}
		})
	}
}

func TestAdEnrichColors(t *testing.T) {
	secStore, err := store.NewSecureStore()
	if err != nil {
		t.Fatalf("Failed to create secure store: %v", err)
	}

	ctrl := NewController(secStore)
	state := PlaybackState{
		IsAd:   true,
		ArtURL: "http://example.com/invalid-image-url",
	}

	ctrl.enrichColors(&state)

	if state.PrimaryColor != "#1db954" {
		t.Errorf("Expected PrimaryColor #1db954 for ad, got %s", state.PrimaryColor)
	}
	if state.SecondaryColor != "#0d1117" {
		t.Errorf("Expected SecondaryColor #0d1117 for ad, got %s", state.SecondaryColor)
	}
}
