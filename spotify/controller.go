package spotify

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/benzjeremy/spotify-screensaver/store"
)

type Controller struct {
	mu           sync.RWMutex
	store        *store.SecureStore
	lastState    PlaybackState
	lastPolled   time.Time
	demoPosition int64
	demoTicker   time.Time
}

func NewController(s *store.SecureStore) *Controller {
	return &Controller{
		store:      s,
		demoTicker: time.Now(),
		lastState: PlaybackState{
			IsPlaying:     false,
			Title:         "Bereit für Musik",
			Artist:        "Spotify",
			Album:         "Screensaver",
			ArtURL:        "",
			DurationMs:    215000,
			VolumePercent: 75,
			Source:        "demo",
			IsConnected:   false,
		},
	}
}

// GetPlaybackState returns the current state of Spotify playback
func (c *Controller) GetPlaybackState() PlaybackState {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Try Linux MPRIS via playerctl
	if state, ok := c.queryMPRIS(); ok {
		c.lastState = state
		c.lastPolled = time.Now()
		return state
	}

	// 2. Try spotify_player CLI
	if state, ok := c.querySpotifyPlayer(); ok {
		c.lastState = state
		c.lastPolled = time.Now()
		return state
	}

	// 3. Fallback demo state with realistic progress simulation
	now := time.Now()
	elapsed := now.Sub(c.demoTicker).Milliseconds()
	c.demoTicker = now
	if c.lastState.IsPlaying {
		c.demoPosition += elapsed
		if c.demoPosition > c.lastState.DurationMs {
			c.demoPosition = 0
		}
	}
	c.lastState.PositionMs = c.demoPosition
	c.lastState.Source = "demo"
	return c.lastState
}

func (c *Controller) queryMPRIS() (PlaybackState, bool) {
	if runtime.GOOS != "linux" {
		return PlaybackState{}, false
	}

	// Try playerctl targeting spotify first, then spotify_player, then any player
	players := []string{"spotify", "spotify_player", "%any"}
	for _, p := range players {
		cmd := exec.Command("playerctl", "-p", p, "metadata", "--format", "{{title}}:::{{artist}}:::{{album}}:::{{mpris:artUrl}}:::{{position}}:::{{mpris:length}}:::{{status}}")
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		parts := strings.Split(strings.TrimSpace(string(out)), ":::")
		if len(parts) < 7 || (parts[0] == "" && parts[1] == "") {
			continue
		}

		title := parts[0]
		artist := parts[1]
		album := parts[2]
		artURL := parts[3]
		posMicros, _ := strconv.ParseInt(parts[4], 10, 64)
		lenMicros, _ := strconv.ParseInt(parts[5], 10, 64)
		status := strings.ToLower(parts[6])

		isPlaying := status == "playing"
		posMs := posMicros / 1000
		durationMs := lenMicros / 1000
		if durationMs <= 0 {
			durationMs = 180000
		}

		// Get volume
		volPercent := 70
		volOut, err := exec.Command("playerctl", "-p", p, "volume").Output()
		if err == nil {
			if f, err := strconv.ParseFloat(strings.TrimSpace(string(volOut)), 64); err == nil {
				volPercent = int(f * 100)
			}
		}

		return PlaybackState{
			IsPlaying:     isPlaying,
			Title:         title,
			Artist:        artist,
			Album:         album,
			ArtURL:        artURL,
			PositionMs:    posMs,
			DurationMs:    durationMs,
			VolumePercent: volPercent,
			PlayerName:    p,
			IsConnected:   true,
			Source:        "mpris",
		}, true
	}

	return PlaybackState{}, false
}

func (c *Controller) querySpotifyPlayer() (PlaybackState, bool) {
	cmd := exec.Command("spotify_player", "get", "key", "playback")
	out, err := cmd.Output()
	if err != nil {
		return PlaybackState{}, false
	}

	var data map[string]interface{}
	if err := json.Unmarshal(out, &data); err != nil {
		return PlaybackState{}, false
	}

	isPlaying, _ := data["is_playing"].(bool)
	progressMs, _ := data["progress_ms"].(float64)

	item, ok := data["item"].(map[string]interface{})
	if !ok {
		return PlaybackState{}, false
	}

	title, _ := item["name"].(string)
	durationMs, _ := item["duration_ms"].(float64)

	var artists []string
	if arts, ok := item["artists"].([]interface{}); ok {
		for _, a := range arts {
			if aMap, ok := a.(map[string]interface{}); ok {
				if aName, ok := aMap["name"].(string); ok {
					artists = append(artists, aName)
				}
			}
		}
	}

	albumName := ""
	artURL := ""
	if alb, ok := item["album"].(map[string]interface{}); ok {
		albumName, _ = alb["name"].(string)
		if images, ok := alb["images"].([]interface{}); ok && len(images) > 0 {
			if img, ok := images[0].(map[string]interface{}); ok {
				artURL, _ = img["url"].(string)
			}
		}
	}

	return PlaybackState{
		IsPlaying:     isPlaying,
		Title:         title,
		Artist:        strings.Join(artists, ", "),
		Album:         albumName,
		ArtURL:        artURL,
		PositionMs:    int64(progressMs),
		DurationMs:    int64(durationMs),
		VolumePercent: 80,
		PlayerName:    "spotify_player",
		IsConnected:   true,
		Source:        "spotify_player",
	}, true
}

// PlayPause toggles current playback
func (c *Controller) PlayPause() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if runtime.GOOS == "linux" {
		exec.Command("playerctl", "-p", "spotify,spotify_player,%any", "play-pause").Run()
	}
	exec.Command("spotify_player", "playback", "play-pause").Run()

	c.lastState.IsPlaying = !c.lastState.IsPlaying
	c.demoTicker = time.Now()
}

// Next skips to next track
func (c *Controller) Next() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if runtime.GOOS == "linux" {
		exec.Command("playerctl", "-p", "spotify,spotify_player,%any", "next").Run()
	}
	exec.Command("spotify_player", "playback", "next").Run()
	c.demoPosition = 0
	c.demoTicker = time.Now()
}

// Previous jumps to previous track
func (c *Controller) Previous() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if runtime.GOOS == "linux" {
		exec.Command("playerctl", "-p", "spotify,spotify_player,%any", "previous").Run()
	}
	exec.Command("spotify_player", "playback", "previous").Run()
	c.demoPosition = 0
	c.demoTicker = time.Now()
}

// SetVolume adjusts volume
func (c *Controller) SetVolume(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastState.VolumePercent += delta
	if c.lastState.VolumePercent < 0 {
		c.lastState.VolumePercent = 0
	}
	if c.lastState.VolumePercent > 100 {
		c.lastState.VolumePercent = 100
	}

	if runtime.GOOS == "linux" {
		volFloat := float64(c.lastState.VolumePercent) / 100.0
		exec.Command("playerctl", "-p", "spotify,spotify_player,%any", "volume", fmt.Sprintf("%.2f", volFloat)).Run()
		sign := "+"
		if delta < 0 {
			sign = "-"
		}
		val := fmt.Sprintf("%d%%", abs(delta))
		exec.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", val+sign).Run()
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Seek jumps to a specific second in the track
func (c *Controller) Seek(seconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.demoPosition = int64(seconds) * 1000
	c.demoTicker = time.Now()

	if runtime.GOOS == "linux" {
		exec.Command("playerctl", "-p", "spotify,spotify_player,%any", "position", strconv.Itoa(seconds)).Run()
	}
}
