package store

import (
	"bytes"
	"testing"
)

func TestCryptoEncryptDecrypt(t *testing.T) {
	secStore, err := NewSecureStore()
	if err != nil {
		t.Fatalf("NewSecureStore failed: %v", err)
	}

	secret := []byte("Jeremy-Benz-Spotify-Token-Secret-12345!#")
	encrypted, err := secStore.Encrypt(secret)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(secret, encrypted) {
		t.Fatalf("Encrypted ciphertext should not match plaintext")
	}

	decrypted, err := secStore.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(secret, decrypted) {
		t.Fatalf("Decrypted %q does not match original %q", decrypted, secret)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	secStore, err := NewSecureStore()
	if err != nil {
		t.Fatalf("NewSecureStore failed: %v", err)
	}

	cfg := &Config{
		SpotifyClientID: "test-client-id",
		RefreshToken:    "test-refresh-token",
		ClockFormat24H:  true,
		VisualizerMode:  "wave",
	}

	if err := secStore.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := secStore.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.SpotifyClientID != cfg.SpotifyClientID {
		t.Errorf("Expected ClientID %s, got %s", cfg.SpotifyClientID, loaded.SpotifyClientID)
	}
	if loaded.VisualizerMode != "wave" {
		t.Errorf("Expected VisualizerMode wave, got %s", loaded.VisualizerMode)
	}
}
