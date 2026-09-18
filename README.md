# 🌌 Spotify Screensaver

[![Go Reference](https://pkg.go.dev/badge/github.com/benzjeremy/spotify-screensaver.svg)](https://pkg.go.dev/github.com/benzjeremy/spotify-screensaver)
[![Go Report Card](https://goreportcard.com/badge/github.com/benzjeremy/spotify-screensaver.svg)](https://goreportcard.com/report/github.com/benzjeremy/spotify-screensaver)
[![CI](https://github.com/benzjeremy/spotify-screensaver/actions/workflows/ci.yml/badge.svg)](https://github.com/benzjeremy/spotify-screensaver/actions)
[![Coverage](https://codecov.io/gh/benzjeremy/spotify-screensaver/branch/main/graph/badge.svg)](https://app.codecov.io/gh/benzjeremy/spotify-screensaver)
[![Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Release](https://img.shields.io/badge/Release-v1.4%20[Pre--Release]-emerald)](https://github.com/benzjeremy/spotify-screensaver/releases/latest)
[![Status: Pre-Release](https://img.shields.io/badge/Status-Pre--Release%20%2F%20WIP-orange.svg)](https://github.com/benzjeremy/spotify-screensaver)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-lightgrey.svg)]()
[![Security: AES-256-GCM](https://img.shields.io/badge/Security-AES--256--GCM-success.svg)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)


> [!IMPORTANT]
> ### 🔒 Primary Codebase & Active Development Moved to Self-Hosted Gitea
> **For privacy optimization and sovereign self-hosted infrastructure, the primary development, source code, and releases of this project have permanently migrated to our self-hosted Gitea platform:**  
> 👉 **[Gitea Repository: https://pi5.darter-basking.ts.net/gitea/spotify-screensaver/spotify-screensaver](https://pi5.darter-basking.ts.net/gitea/spotify-screensaver/spotify-screensaver)**  
> 👉 **[Official Web Showcase: https://pi5.darter-basking.ts.net/spotify-screensaver/](https://pi5.darter-basking.ts.net/spotify-screensaver/)**
> 
> *This GitHub repository serves solely as a read-only mirror for Go toolchain compatibility (`go install`, `pkg.go.dev`, `awesome-go`). All active development, issues, and releases take place on Gitea.*

---

> [!IMPORTANT]
> ### 🚧 Pre-Release / Active Development Notice
> **This software is not yet finished and is under active development.**  
> All releases and binaries are **Pre-Releases** (Work in Progress), even if originally tagged or announced without a pre-release flag. Features, hardware tap APIs, and UI designs are actively developed and continually updated.

> An elegant, resource-friendly desktop screensaver with **Live OLED clock**, **Spotify song metadata**, **intelligent ad handling**, **focused clean design**, **hardware-coupled audio spectrum analysis (PulseAudio/PipeWire)**, **dynamic cover color palette**, **circular visualizer**, and an **inactivity daemon** for Linux (WebKitGTK) and Windows (App Mode). **100% Local-First & Zero Bloat.**

---

## ✨ Features & Highlights in v1.4

- 📢 **Intelligent Ad Detection & Handling (Spotify Free):**
  - Detects advertisements via MPRIS track IDs (`:ad:`), metadata, and API payloads.
  - Automatically switches to an embedded dark-themed ad placeholder (`ad-placeholder.svg`) with a subtle pulsating "ADVERTISEMENT" badge.
  - Clean fallback metadata ("Spotify Advertisement" / "Commercial Break") instead of frozen stale cover art or rendering glitches.
- 🎨 **Focused Clean Layout (Redundancy-Free Cover Presentation):**
  - Completely replaced the redundant vinyl/CD visual with a centered album art aesthetic featuring soft ambient glow and rounded geometry.
- 🎛️ **True Hardware Audio Spectrum Analysis (PulseAudio / PipeWire / WASAPI):**
  - Direct hardware tap of system audio via monitor sink.
  - 64-band FFT analysis (Cooley-Tukey Radix-2) with 60 FPS WebSocket streaming (`/api/audio-stream`).
  - Zero-latency visualization of real bass, mid, and treble frequencies rather than simulation.
- 🎨 **Dynamic Cover Palette (Adaptive Glow & Accents):**
  - Fast color space reduction and dominance extraction (K-Means / quantization) in Go.
  - Smooth 1.2s CSS/Canvas color transitions adapting screensaver glow and accents to the album's primary and secondary palette.
- 🌊 **Circular Visualizer Mode:**
  - 360° radial frequency spectrum around album art with glowing peak highlights.
- 💤 **Idle Detection Daemon (`--idle-timeout=N`):**
  - Monitors user inactivity via `xprintidle` / D-Bus / Win32 and activates the screensaver automatically.
- 🕒 **OLED Digital Clock & Date:**
  - High-contrast neon time display with configurable seconds toggle and 12h/24h format support.
- 🎨 **Color Accents & Themes (4 Styles):**
  - 🟢 **Spotify Classic:** Signature Spotify green with soft ambient neon.
  - 🔷 **Electric Cyan:** Futuristic ice-blue neon.
  - 🟣 **Neon Purple:** Cyberpunk violet / deep magenta.
  - 🟡 **Sunset Amber:** Warm gold / amber tone.
- 🌊 **Multi-Mode Canvas Audio Visualizer:**
  - **Bars:** 48 frequency bars with physical peak-hold drops and gradients.
  - **Wave:** Flowing oscilloscope wave with soft neon blur.
  - **Mirrored:** Mirrored dual columns radiating from the center.
  - Configurable sensitivity and responsiveness.
- 🎛️ **Interactive Playback Controls:**
  - **Progress Scrubber:** Click anywhere on the track bar to seek.
  - **Volume Slider:** Seamless volume adjustment via slider or keyboard shortcuts.
- 🎵 **Zero-Config Spotify MPRIS:**
  - Automatically detects Spotify Desktop and `spotify_player` on Linux over D-Bus without requiring API keys.
  - Smooth standby demo mode when Spotify is paused or closed.
- 🛡️ **Strict Security Architecture (Jeremy Benz Standards):**
  - **Cryptography:** AES-256-GCM token encryption derived via PBKDF2 (100,000 rounds, hardware fingerprint, unique salt stored in `~/.config/spotify-screensaver/salt.bin`).
  - **Network Isolation:** Local HTTP server binds strictly to `127.0.0.1:43210`.
  - **Anti-DNS-Rebinding & Anti-CSRF:** Strict validation of `Host` and `Origin` headers.
  - **Security Headers:** `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, strict CSP.
- 💤 **Automatic Inactivity Fade:** Hides mouse cursor and HUD overlays after 4 seconds of idle time.

---

## ⌨️ Keyboard Shortcuts

| Key | Action |
|---|---|
| <kbd>Space</kbd> | Play / Pause |
| <kbd>→</kbd> | Next Track |
| <kbd>←</kbd> | Previous Track |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Volume +5% / -5% |
| <kbd>F</kbd> or <kbd>F11</kbd> | Toggle Fullscreen Mode |
| <kbd>ESC</kbd> | Exit Screensaver / Close Settings Modal |

---

## 🚀 Installation & Usage

### 1. Build from Source (Linux with WebKitGTK)

```bash
cd ~/Projekte/benzjeremy.github.io/spotify-screensaver
go build -o spotify-screensaver .
./spotify-screensaver
```

### 2. Run in Fullscreen Screensaver Mode

```bash
./spotify-screensaver -fullscreen
```

### 3. Run in Default Browser Mode

```bash
./spotify-screensaver -browser
```

### 4. Cross-Compile for Windows

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/spotify-screensaver-windows-amd64.exe .
```

---

## ⚖️ License & Author

- **Developer:** Jeremy Benz ([@benzjeremy](https://github.com/benzjeremy)) · [benzjeremy.github.io](https://benzjeremy.github.io/)
- **License:** [GNU General Public License v3.0 (GPL-3.0)](LICENSE)