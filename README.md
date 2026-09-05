# 🌌 Spotify Screensaver

[![Go Reference](https://pkg.go.dev/badge/github.com/benzjeremy/spotify-screensaver.svg)](https://pkg.go.dev/github.com/benzjeremy/spotify-screensaver)
[![Go Report Card](https://goreportcard.com/badge/github.com/benzjeremy/spotify-screensaver.svg)](https://goreportcard.com/report/github.com/benzjeremy/spotify-screensaver)
[![CI](https://github.com/benzjeremy/spotify-screensaver/actions/workflows/ci.yml/badge.svg)](https://github.com/benzjeremy/spotify-screensaver/actions)
[![Coverage](https://codecov.io/gh/benzjeremy/spotify-screensaver/branch/main/graph/badge.svg)](https://app.codecov.io/gh/benzjeremy/spotify-screensaver)
[![Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Release](https://img.shields.io/badge/Release-Latest-emerald)](https://github.com/benzjeremy/spotify-screensaver/releases/latest)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-lightgrey.svg)]()

> Ein eleganter, ressourcenschonender Desktop-Bildschirmschoner mit **Live-OLED-Uhrzeit**, **Spotify Song-Metadaten**, **hardwarebeschleunigtem Multi-Mode HTML5 Canvas Visualizer** und **Musiksteuerung** für Linux (WebKitGTK) und Windows (App-Mode). **100% Local-First & Zero Bloat (Kein KI-Overhead).**

---

## ✨ Features in v1.1 (Hotfix & Layout Upgrade)

- 🕒 **OLED Digitaluhr & deutsches Datum:**
  - Große Neon-Zeitanzeige mit konfigurierbarer Sekundenanzeige (ein-/ausblendbar) und 12h/24h-Umschaltung.
  - Deutsches Datumsformat (z. B. "Samstag, 5. September 2026").
- 🎨 **Farbakzente & Themes (4 Stile):**
  - 🟢 **Spotify Classic:** Echtes Spotify-Grün mit feinem Glow.
  - 🔷 **Electric Cyan:** Futuristisches Ice-Blue Neon.
  - 🟣 **Neon Purple:** Cyberpunk Violett / Deep Magenta.
  - 🟡 **Sunset Amber:** Warmer Gold-/Amber-Ton.
- 🌊 **Multi-Mode Canvas Audio-Visualizer:**
  - **Bars:** 48 Frequenzbalken mit physikalischem Peak-Hold-Drop und Farbverlauf.
  - **Wave:** Fließende Oszillograph-Welle mit weichem Glow.
  - **Mirrored:** Gespiegelte Doppelsäulen vom Zentrum aus.
  - Regulierbare Sensitivität / Reaktionsstärke.
- 🎛️ **Interaktiver Player & Steuerung:**
  - **Progress Scrubber:** Durch Klick auf die Fortschrittsleiste kann direkt zu jeder Songposition gesprungen werden.
  - **Volume-Slider:** Stufenlose Lautstärkeregelung direkt per Slider oder Tastatur.
  - **Vinyl-Schallplatten-Animation:** Sanft rotierende Schallplatte mit metallischem Groove-Glanz bei aktiver Wiedergabe.
- 🎵 **Zero-Config Spotify MPRIS:**
  - Erkennt Spotify Desktop und `spotify_player` unter Linux ohne API-Key direkt über D-Bus.
  - Flüssiger Demo-/Standby-Modus, wenn Spotify pausiert oder geschlossen ist.
- 🛡️ **Verbindliche Sicherheitsarchitektur (Jeremy Benz Standards):**
  - **Kryptografie:** AES-256-GCM Token-Verschlüsselung, abgeleitet via PBKDF2 (100.000 Runden, Hardware-Fingerprint, unikat generiertes Salt in `~/.config/spotify-screensaver/salt.bin`).
  - **Netzwerksicherheit:** Lokaler HTTP-Server bindet strikt an `127.0.0.1:43210`.
  - **Anti-DNS-Rebinding & Anti-CSRF:** Strikte Validierung des `Host`- und `Origin`-Headers.
  - **Security Headers:** `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, CSP.
- 💤 **Screensaver-Automatik:** Automatisches Ausblenden von Mauszeiger und HUD nach 4 Sekunden Inaktivität.

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

## ⚖️ Lizenz

Dieses Projekt ist unter der **GNU General Public License v3.0 (GPL-3.0)** lizenziert – siehe [LICENSE](LICENSE) für Details.

Entwickelt von **Jeremy Benz** · [benzjeremy.github.io](https://benzjeremy.github.io/)
