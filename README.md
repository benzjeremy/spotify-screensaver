# 🌌 Spotify Screensaver

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-lightgrey.svg)]()
[![Security](https://img.shields.io/badge/Security-AES--256--GCM%20%7C%20PBKDF2-green.svg)]()

> Ein eleganter, nativer Desktop-Bildschirmschoner mit **Live-Uhrzeit**, **Spotify Song-Metadaten**, **HTML5 Canvas Audio-Visualizer** und **integrierter Musiksteuerung** für Linux (WebKitGTK) und Windows (App-Mode).

---

## ✨ Features

- 🕒 **OLED Digitaluhr & Datum:** Große Neon-Zeitanzeige (HH:MM:SS) mit Sekundenanzeige und lokalisiertem deutschen Datum.
- 🎵 **Spotify Live-Integration:**
  - Zero-Config Linux MPRIS / D-Bus Unterstützung (erkennt Spotify Desktop & `spotify_player` sofort ohne API-Key).
  - Albumcover, Titel, Artist, Albumname und Live-Fortschrittsanzeige mit rotierender Vinyl-Schallplatten-Animation.
- 🎛️ **Interaktive Steuerung:** Play/Pause, Next Track, Previous Track, Lautstärkeregelung via Maus und Tastatur-Shortcuts.
- 🌊 **Dynamic Canvas Spectrum Visualizer:** 48 Frequenzbänder mit physikalischer Spitzenwert-Dämpfung (Peak Caps) und Neon-Glow.
- 🛡️ **Verbindliche Sicherheitsarchitektur (Jeremy Benz Standards):**
  - **Kryptografie:** Secrets & Tokens werden mit **AES-256-GCM** verschlüsselt, abgeleitet via **PBKDF2** mit mindestens **100.000 Iterationen** und hardware-gebundenem Salt.
  - **Netzwerk:** Lokaler HTTP-Server bindet ausschließlich an `127.0.0.1:43210`.
  - **Anti-DNS-Rebinding & Anti-CSRF:** Strikte Validierung des `Host`- und `Origin`-Headers.
  - **Security Headers:** `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, Content Security Policy (CSP).
- 💤 **Screensaver-Automatik:** Automatisches Ausblenden des Mauszeigers und HUD nach 4 Sekunden Inaktivität.
- 🤖 **KI Mood-DJ (Optional):** Integration für FreeLLM / OpenAI zur automatischen Song-Filterung ("Auto-Skip Depri").

---

## ⌨️ Tastatur-Shortcuts

| Taste | Aktion |
|---|---|
| <kbd>Space</kbd> | Wiedergabe / Pause |
| <kbd>→</kbd> | Nächster Song (Next) |
| <kbd>←</kbd> | Vorheriger Song (Previous) |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Lautstärke +5% / -5% |
| <kbd>F</kbd> oder <kbd>F11</kbd> | Vollbildmodus umschalten |
| <kbd>ESC</kbd> | Screensaver beenden / Einstellungen schließen |

---

## 🚀 Installation & Start

### 1. Aus Quellcode kompilieren (Linux mit WebKitGTK)

```bash
# Voraussetzungen auf Arch / CachyOS:
# webkit2gtk-4.1 gtk3 playerctl

cd ~/Projekte/benzjeremy.github.io/spotify-screensaver
go build -o spotify-screensaver .
./spotify-screensaver
```

### 2. Als Vollbild-Screensaver starten

```bash
./spotify-screensaver -fullscreen
```

### 3. Im Standard-Browser starten

```bash
./spotify-screensaver -browser
```

### 4. Cross-Compilation für Windows

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/spotify-screensaver-windows-amd64.exe .
```

---

## 🏛️ Architektur

```
spotify-screensaver/
├── assets/                  # Embedded UI (HTML5, CSS3, ES6 JS, Canvas)
│   ├── index.html
│   ├── style.css
│   └── app.js
├── store/                   # AES-256-GCM & PBKDF2 (100k) Token-Speicher
│   ├── crypto.go
│   └── crypto_test.go
├── spotify/                 # MPRIS D-Bus & Spotify Player Integration
│   ├── controller.go
│   └── types.go
├── server/                  # 127.0.0.1 Server, Anti-CSRF, Anti-DNS-Rebinding
│   ├── server.go
│   └── server_test.go
├── gui_linux.go             # WebKitGTK Native Window (Linux)
├── gui_windows.go           # Chrome/Edge App-Mode Launcher (Windows)
├── gui_other.go             # Fallback Launcher
└── main.go                  # Einstiegspunkt & Lifecycle-Management
```

---

## ⚖️ Lizenz

Dieses Projekt ist unter der **GNU General Public License v3.0 (GPL-3.0)** lizenziert – siehe [LICENSE](LICENSE) für Details.

Entwickelt von **Jeremy Benz** · [benzjeremy.github.io](https://benzjeremy.github.io/)
