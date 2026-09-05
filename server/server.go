package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/benzjeremy/spotify-screensaver/spotify"
	"github.com/benzjeremy/spotify-screensaver/store"
)

type Server struct {
	port       int
	ctrl       *spotify.Controller
	store      *store.SecureStore
	assets     fs.FS
	listener   net.Listener
	httpServer *http.Server
	mu         sync.Mutex
}

func NewServer(port int, ctrl *spotify.Controller, store *store.SecureStore, assets fs.FS) *Server {
	return &Server{
		port:   port,
		ctrl:   ctrl,
		store:  store,
		assets: assets,
	}
}

func (s *Server) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// If preferred port is in use, fallback to any free loopback port
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", fmt.Errorf("failed to bind 127.0.0.1: %w", err)
		}
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()

	// Static Assets
	fileServer := http.FileServer(http.FS(s.assets))
	mux.Handle("/", fileServer)

	// API Endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/playpause", s.handlePlayPause)
	mux.HandleFunc("/api/next", s.handleNext)
	mux.HandleFunc("/api/previous", s.handlePrevious)
	mux.HandleFunc("/api/seek", s.handleSeek)
	mux.HandleFunc("/api/volume", s.handleVolume)
	mux.HandleFunc("/api/config", s.handleConfig)

	// Wrap with security middleware
	handler := s.securityMiddleware(mux)

	s.httpServer = &http.Server{
		Handler: handler,
	}

	go s.httpServer.Serve(listener)

	url := fmt.Sprintf("http://127.0.0.1:%d", s.port)
	return url, nil
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Anti-DNS-Rebinding: Validate Host header strictly
		host := r.Host
		if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
			host = host[:colonIdx]
		}
		if host != "127.0.0.1" && host != "localhost" {
			http.Error(w, "Forbidden: Anti-DNS-Rebinding check failed", http.StatusForbidden)
			return
		}

		// 2. Anti-CSRF on mutation requests
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if !strings.HasPrefix(origin, "http://127.0.0.1:") && !strings.HasPrefix(origin, "http://localhost:") {
					http.Error(w, "Forbidden: Invalid origin", http.StatusForbidden)
					return
				}
			}
		}

		// 3. Security Headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' https: data:; img-src 'self' data: https: http:; media-src 'self' https:;")

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state := s.ctrl.GetPlaybackState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (s *Server) handlePlayPause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.ctrl.PlayPause()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.ctrl.Next()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePrevious(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.ctrl.Previous()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleVolume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Delta int `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	s.ctrl.SetVolume(req.Delta)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Seconds int `json:"seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	s.ctrl.Seek(req.Seconds)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := s.store.LoadConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		publicCfg := map[string]interface{}{
			"clock_format_24h": cfg.ClockFormat24H,
			"show_seconds":    cfg.ShowSeconds,
			"visualizer_mode": cfg.VisualizerMode,
			"theme_accent":    cfg.ThemeAccent,
			"sensitivity":     cfg.Sensitivity,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(publicCfg)

	case http.MethodPost:
		var incoming struct {
			ClockFormat24H bool   `json:"clock_format_24h"`
			ShowSeconds    bool   `json:"show_seconds"`
			VisualizerMode string `json:"visualizer_mode"`
			ThemeAccent    string `json:"theme_accent"`
			Sensitivity    int    `json:"sensitivity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		cfg, _ := s.store.LoadConfig()
		cfg.ClockFormat24H = incoming.ClockFormat24H
		cfg.ShowSeconds = incoming.ShowSeconds
		if incoming.VisualizerMode != "" {
			cfg.VisualizerMode = incoming.VisualizerMode
		}
		if incoming.ThemeAccent != "" {
			cfg.ThemeAccent = incoming.ThemeAccent
		}
		if incoming.Sensitivity > 0 {
			cfg.Sensitivity = incoming.Sensitivity
		}

		if err := s.store.SaveConfig(cfg); err != nil {
			http.Error(w, "Failed to save secure config", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) Stop() {
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}
