package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benzjeremy/spotify-screensaver/spotify"
	"github.com/benzjeremy/spotify-screensaver/store"
)

func TestSecurityMiddlewareAntiDNSRebinding(t *testing.T) {
	secStore, err := store.NewSecureStore()
	if err != nil {
		t.Fatalf("NewSecureStore error: %v", err)
	}

	ctrl := spotify.NewController(secStore)
	srv := NewServer(0, ctrl, secStore, nil)
	handler := srv.securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with malicious external Host header
	req := httptest.NewRequest("GET", "http://evil-attacker.com/api/status", nil)
	req.Host = "evil-attacker.com"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden on foreign Host header, got %d", rec.Code)
	}

	// Test with valid loopback Host header
	reqValid := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status", nil)
	reqValid.Host = "127.0.0.1:43210"
	recValid := httptest.NewRecorder()
	handler.ServeHTTP(recValid, reqValid)

	if recValid.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on valid 127.0.0.1 Host header, got %d", recValid.Code)
	}

	// Check Security Headers
	if recValid.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Missing X-Frame-Options header")
	}
	if recValid.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Missing X-Content-Type-Options header")
	}
	if !strings.Contains(recValid.Header().Get("Content-Security-Policy"), "default-src") {
		t.Errorf("Missing CSP header")
	}
}
