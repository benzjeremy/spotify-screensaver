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

	// Test with valid loopback Host header and valid Token
	reqValid := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status", nil)
	reqValid.Host = "127.0.0.1:43210"
	reqValid.Header.Set("X-Session-Token", srv.GetSessionToken())
	recValid := httptest.NewRecorder()
	handler.ServeHTTP(recValid, reqValid)

	if recValid.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on valid 127.0.0.1 Host header with token, got %d", recValid.Code)
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

func TestSecurityMiddlewareTokenAuth(t *testing.T) {
	secStore, err := store.NewSecureStore()
	if err != nil {
		t.Fatalf("NewSecureStore error: %v", err)
	}

	ctrl := spotify.NewController(secStore)
	srv := NewServer(0, ctrl, secStore, nil)
	handler := srv.securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Missing Token on /api/ -> 401
	reqNoToken := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status", nil)
	reqNoToken.Host = "127.0.0.1:43210"
	recNoToken := httptest.NewRecorder()
	handler.ServeHTTP(recNoToken, reqNoToken)
	if recNoToken.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized without token, got %d", recNoToken.Code)
	}

	// 2. Invalid Token on /api/ -> 401
	reqBadToken := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status?token=wrong", nil)
	reqBadToken.Host = "127.0.0.1:43210"
	recBadToken := httptest.NewRecorder()
	handler.ServeHTTP(recBadToken, reqBadToken)
	if recBadToken.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized with invalid token, got %d", recBadToken.Code)
	}

	// 3. Valid Token via Query -> 200
	reqQueryToken := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status?token="+srv.GetSessionToken(), nil)
	reqQueryToken.Host = "127.0.0.1:43210"
	recQueryToken := httptest.NewRecorder()
	handler.ServeHTTP(recQueryToken, reqQueryToken)
	if recQueryToken.Code != http.StatusOK {
		t.Errorf("Expected 200 OK with valid query token, got %d", recQueryToken.Code)
	}

	// 4. Valid Token via Header -> 200
	reqHeaderToken := httptest.NewRequest("GET", "http://127.0.0.1:43210/api/status", nil)
	reqHeaderToken.Host = "127.0.0.1:43210"
	reqHeaderToken.Header.Set("X-Session-Token", srv.GetSessionToken())
	recHeaderToken := httptest.NewRecorder()
	handler.ServeHTTP(recHeaderToken, reqHeaderToken)
	if recHeaderToken.Code != http.StatusOK {
		t.Errorf("Expected 200 OK with valid header token, got %d", recHeaderToken.Code)
	}
}
