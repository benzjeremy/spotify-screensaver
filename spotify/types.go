package spotify

type PlaybackState struct {
	IsPlaying     bool   `json:"is_playing"`
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	Album         string `json:"album"`
	ArtURL        string `json:"art_url"`
	PositionMs    int64  `json:"position_ms"`
	DurationMs    int64  `json:"duration_ms"`
	VolumePercent int    `json:"volume_percent"`
	PlayerName    string `json:"player_name"`
	IsConnected   bool   `json:"is_connected"`
	Source        string `json:"source"` // "mpris", "spotify_player", "web_api", "demo"
}
