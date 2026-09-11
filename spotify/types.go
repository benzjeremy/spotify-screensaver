package spotify

type PlaybackState struct {
	IsPlaying     bool   `json:"is_playing"`
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	Album         string `json:"album"`
	ArtURL         string `json:"art_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	PositionMs     int64  `json:"position_ms"`
	DurationMs    int64  `json:"duration_ms"`
	VolumePercent int    `json:"volume_percent"`
	PlayerName    string `json:"player_name"`
	IsConnected   bool   `json:"is_connected"`
	Source        string `json:"source"` // "mpris", "spotify_player", "web_api", "demo"
	IsAd          bool   `json:"is_ad"`
	AdTitle       string `json:"ad_title"`
	AdCoverURL    string `json:"ad_cover_url"`
}
