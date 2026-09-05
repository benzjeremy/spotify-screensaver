package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 100000
	keyLen           = 32 // 256-bit AES
	saltLen          = 32 // 256-bit Salt
)

type Config struct {
	SpotifyClientID     string `json:"spotify_client_id"`
	SpotifyClientSecret string `json:"spotify_client_secret"`
	RefreshToken        string `json:"refresh_token"`
	AccessToken         string `json:"access_token"`
	TokenExpiresAt      int64  `json:"token_expires_at"`
	AIBaseURL           string `json:"ai_base_url"`
	AIAPIKey            string `json:"ai_api_key"`
	AIModel             string `json:"ai_model"`
	AutoSkipDepri       bool   `json:"auto_skip_depri"`
	ClockFormat24H      bool   `json:"clock_format_24h"`
	VisualizerMode      string `json:"visualizer_mode"`
}

type SecureStore struct {
	mu        sync.RWMutex
	configDir string
	saltPath  string
	storePath string
	key       []byte
}

func NewSecureStore() (*SecureStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir error: %w", err)
	}

	configDir := filepath.Join(home, ".config", "spotify-screensaver")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create config dir: %w", err)
	}

	saltPath := filepath.Join(configDir, "salt.bin")
	storePath := filepath.Join(configDir, "secrets.enc")

	salt, err := getOrCreateSalt(saltPath)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize salt: %w", err)
	}

	passphrase := getDevicePassphrase()
	derivedKey := pbkdf2.Key([]byte(passphrase), salt, pbkdf2Iterations, keyLen, sha256.New)

	return &SecureStore{
		configDir: configDir,
		saltPath:  saltPath,
		storePath: storePath,
		key:       derivedKey,
	}, nil
}

func getOrCreateSalt(path string) ([]byte, error) {
	if data, err := os.ReadFile(path); err == nil && len(data) == saltLen {
		return data, nil
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, salt, 0600); err != nil {
		return nil, err
	}
	return salt, nil
}

func getDevicePassphrase() string {
	// Derive stable hardware-bound passphrase without plain-text storage
	identifiers := []string{runtime.GOOS, runtime.GOARCH}

	if machineID, err := os.ReadFile("/etc/machine-id"); err == nil {
		identifiers = append(identifiers, string(machineID))
	} else if hostID, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		identifiers = append(identifiers, string(hostID))
	}

	if hostname, err := os.Hostname(); err == nil {
		identifiers = append(identifiers, hostname)
	}

	combined := fmt.Sprintf("%v-spotify-screensaver-secure-v1", identifiers)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// Encrypt encrypts plaintext bytes with AES-256-GCM using standard 12-byte nonce
func (s *SecureStore) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext bytes with AES-256-GCM
func (s *SecureStore) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption authentication failed: %w", err)
	}

	return plaintext, nil
}

// SaveConfig encrypts and stores the configuration
func (s *SecureStore) SaveConfig(cfg *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	encrypted, err := s.Encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.storePath, encrypted, 0600)
}

// LoadConfig decrypts and loads the configuration
func (s *SecureStore) LoadConfig() (*Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	defaultConfig := &Config{
		ClockFormat24H: true,
		VisualizerMode: "bars",
		AIBaseURL:      "http://localhost:9001",
		AIModel:        "auto",
		AutoSkipDepri:  false,
	}

	encrypted, err := os.ReadFile(s.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, nil
		}
		return nil, err
	}

	decrypted, err := s.Decrypt(encrypted)
	if err != nil {
		// If decryption fails (e.g. initial setup or key mismatch), return defaults
		return defaultConfig, nil
	}

	cfg := &Config{}
	if err := json.Unmarshal(decrypted, cfg); err != nil {
		return defaultConfig, nil
	}

	return cfg, nil
}
