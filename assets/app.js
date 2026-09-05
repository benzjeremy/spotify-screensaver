(() => {
  // DOM Elements
  const hoursEl = document.getElementById("hours");
  const minutesEl = document.getElementById("minutes");
  const secondsEl = document.getElementById("seconds");
  const dateDisplayEl = document.getElementById("dateDisplay");
  const ambientGlowEl = document.getElementById("ambientGlow");

  const playerCardEl = document.getElementById("playerCard");
  const coverImgEl = document.getElementById("coverImg");
  const trackTitleEl = document.getElementById("trackTitle");
  const trackArtistEl = document.getElementById("trackArtist");
  const trackAlbumEl = document.getElementById("trackAlbum");
  const progressBarContainer = document.getElementById("progressBarContainer");
  const progressFillEl = document.getElementById("progressFill");
  const progressHandleEl = document.getElementById("progressHandle");
  const currentTimeEl = document.getElementById("currentTime");
  const totalTimeEl = document.getElementById("totalTime");

  const btnPrev = document.getElementById("btnPrev");
  const btnPlay = document.getElementById("btnPlay");
  const btnNext = document.getElementById("btnNext");
  const playIcon = document.getElementById("playIcon");
  const pauseIcon = document.getElementById("pauseIcon");
  const volSlider = document.getElementById("volSlider");
  const volTextEl = document.getElementById("volText");

  const sourceLabelEl = document.getElementById("sourceLabel");
  const statusDotEl = document.getElementById("statusDot");
  const hudFooterEl = document.getElementById("hudFooter");

  const canvas = document.getElementById("visualizerCanvas");
  const ctx = canvas.getContext("2d");

  const btnSettings = document.getElementById("btnSettings");
  const modalSettings = document.getElementById("modalSettings");
  const btnCloseModal = document.getElementById("btnCloseModal");
  const btnSaveConfig = document.getElementById("btnSaveConfig");
  const cfgVisualizer = document.getElementById("cfgVisualizer");
  const cfgSensitivity = document.getElementById("cfgSensitivity");
  const cfgClockFormat = document.getElementById("cfgClockFormat");
  const cfgShowSeconds = document.getElementById("cfgShowSeconds");
  const themeSwatches = document.querySelectorAll(".btn-theme-swatch");

  // State
  let currentState = {
    is_playing: false,
    title: "Bereit für Musik",
    artist: "Spotify",
    album: "Screensaver",
    art_url: "",
    position_ms: 0,
    duration_ms: 215000,
    volume_percent: 75,
    source: "demo"
  };

  let localPositionMs = 0;
  let lastSyncTime = Date.now();
  let clock24H = true;
  let showSeconds = true;
  let visualizerMode = "bars";
  let visualizerSensitivity = 80;
  let currentTheme = "spotify";

  // Inactivity auto-hide
  let mouseTimer = null;
  function resetInactivity() {
    document.body.classList.remove("hide-cursor");
    hudFooterEl.classList.remove("faded");
    clearTimeout(mouseTimer);
    mouseTimer = setTimeout(() => {
      document.body.classList.add("hide-cursor");
      hudFooterEl.classList.add("faded");
    }, 4000);
  }
  window.addEventListener("mousemove", resetInactivity);
  window.addEventListener("keydown", resetInactivity);
  resetInactivity();

  // 1. Clock & Date Update
  function updateClock() {
    const now = new Date();
    let h = now.getHours();
    if (!clock24H) {
      h = h % 12 || 12;
    }
    const m = String(now.getMinutes()).padStart(2, "0");
    const s = String(now.getSeconds()).padStart(2, "0");

    hoursEl.textContent = String(h).padStart(2, "0");
    minutesEl.textContent = m;
    if (showSeconds) {
      secondsEl.style.display = "inline";
      secondsEl.textContent = s;
    } else {
      secondsEl.style.display = "none";
    }

    // German Date format
    const options = { weekday: "long", day: "numeric", month: "long", year: "numeric" };
    dateDisplayEl.textContent = now.toLocaleDateString("de-DE", options);
  }
  setInterval(updateClock, 200);
  updateClock();

  // 2. Playback State Polling & Interpolation
  function formatMs(ms) {
    if (!ms || ms < 0) return "0:00";
    const totalSec = Math.floor(ms / 1000);
    const m = Math.floor(totalSec / 60);
    const s = String(totalSec % 60).padStart(2, "0");
    return `${m}:${s}`;
  }

  async function pollStatus() {
    try {
      const res = await fetch("/api/status");
      if (res.ok) {
        const data = await res.json();
        applyPlaybackState(data);
      }
    } catch (err) {
      // Backend temporarily unavailable
    }
  }

  function applyPlaybackState(data) {
    const wasPlaying = currentState.is_playing;
    currentState = data;
    localPositionMs = data.position_ms;
    lastSyncTime = Date.now();

    trackTitleEl.textContent = data.title || "Unbekannter Titel";
    trackArtistEl.textContent = data.artist || "Spotify";
    trackAlbumEl.textContent = data.album || "";
    volTextEl.textContent = `${data.volume_percent}%`;
    volSlider.value = data.volume_percent;

    totalTimeEl.textContent = formatMs(data.duration_ms);

    if (data.art_url && data.art_url.startsWith("http")) {
      coverImgEl.src = data.art_url;
    }

    if (data.is_playing) {
      playerCardEl.classList.add("is-playing");
      playIcon.classList.add("hidden");
      pauseIcon.classList.remove("hidden");
    } else {
      playerCardEl.classList.remove("is-playing");
      playIcon.classList.remove("hidden");
      pauseIcon.classList.add("hidden");
    }

    // Source indicator
    if (data.source === "mpris") {
      sourceLabelEl.textContent = `Spotify MPRIS (${data.player_name})`;
      statusDotEl.style.opacity = "1";
    } else if (data.source === "spotify_player") {
      sourceLabelEl.textContent = "spotify_player CLI";
      statusDotEl.style.opacity = "1";
    } else {
      sourceLabelEl.textContent = "Demo / Standby";
      statusDotEl.style.opacity = "0.4";
    }
  }

  // Throttled local timeline ticker (10 FPS DOM updates - reduces CPU by >80%)
  let lastTickTime = 0;
  function tickTimeline(now) {
    if (!lastTickTime || now - lastTickTime >= 100) {
      lastTickTime = now;
      if (currentState.is_playing && currentState.duration_ms > 0) {
        const delta = Date.now() - lastSyncTime;
        const currentPos = Math.min(localPositionMs + delta, currentState.duration_ms);
        currentTimeEl.textContent = formatMs(currentPos);
        const pct = (currentPos / currentState.duration_ms) * 100;
        progressFillEl.style.width = `${pct}%`;
        progressHandleEl.style.left = `${pct}%`;
      } else {
        currentTimeEl.textContent = formatMs(localPositionMs);
        const pct = currentState.duration_ms > 0 ? (localPositionMs / currentState.duration_ms) * 100 : 0;
        progressFillEl.style.width = `${pct}%`;
        progressHandleEl.style.left = `${pct}%`;
      }
    }
    requestAnimationFrame(tickTimeline);
  }
  requestAnimationFrame(tickTimeline);

  setInterval(pollStatus, 1000);
  pollStatus();

  // 3. Controls API
  async function postAction(endpoint, body = {}) {
    try {
      await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
      });
      setTimeout(pollStatus, 150);
    } catch (e) {
      console.error(e);
    }
  }

  btnPlay.addEventListener("click", () => postAction("/api/playpause"));
  btnNext.addEventListener("click", () => postAction("/api/next"));
  btnPrev.addEventListener("click", () => postAction("/api/previous"));

  // Interactive Progress Scrubber Click
  progressBarContainer.addEventListener("click", (e) => {
    if (currentState.duration_ms <= 0) return;
    const rect = progressBarContainer.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const pct = Math.max(0, Math.min(1, clickX / rect.width));
    const targetMs = Math.floor(pct * currentState.duration_ms);
    const targetSec = Math.floor(targetMs / 1000);
    localPositionMs = targetMs;
    lastSyncTime = Date.now();
    postAction("/api/seek", { seconds: targetSec });
  });

  // Interactive Volume Slider
  volSlider.addEventListener("input", (e) => {
    const val = parseInt(e.target.value, 10);
    volTextEl.textContent = `${val}%`;
    const delta = val - currentState.volume_percent;
    if (delta !== 0) {
      postAction("/api/volume", { delta });
    }
  });

  // 4. Keyboard Shortcuts
  window.addEventListener("keydown", (e) => {
    // Ignore when typing inside modal inputs
    if (e.target.tagName === "INPUT" || e.target.tagName === "SELECT") return;

    switch (e.code) {
      case "Space":
        e.preventDefault();
        postAction("/api/playpause");
        break;
      case "ArrowRight":
        e.preventDefault();
        postAction("/api/next");
        break;
      case "ArrowLeft":
        e.preventDefault();
        postAction("/api/previous");
        break;
      case "ArrowUp":
        e.preventDefault();
        postAction("/api/volume", { delta: 5 });
        break;
      case "ArrowDown":
        e.preventDefault();
        postAction("/api/volume", { delta: -5 });
        break;
      case "KeyF":
        if (!document.fullscreenElement) {
          document.documentElement.requestFullscreen().catch(() => {});
        } else {
          document.exitFullscreen().catch(() => {});
        }
        break;
      case "Escape":
        if (!modalSettings.classList.contains("hidden")) {
          modalSettings.classList.add("hidden");
        } else {
          // If in fullscreen, exit fullscreen first
          if (document.fullscreenElement) {
            document.exitFullscreen().catch(() => {});
          }
        }
        break;
    }
  });

  // 5. Canvas Audio Visualizer
  const numBars = 48;
  const bars = [];
  for (let i = 0; i < numBars; i++) {
    bars.push({
      height: 10,
      targetHeight: 10,
      peak: 10,
      peakHold: 0,
      speed: 0.1 + Math.random() * 0.15
    });
  }

  // Theme definitions (cached to avoid getComputedStyle layout thrashing)
  const THEME_CONFIG = {
    spotify: { primary: "#1db954", glow: "rgba(29, 185, 84, 0.4)" },
    cyan: { primary: "#00f2ff", glow: "rgba(0, 242, 255, 0.4)" },
    purple: { primary: "#c084fc", glow: "rgba(192, 132, 252, 0.45)" },
    amber: { primary: "#fbbf24", glow: "rgba(251, 191, 36, 0.45)" }
  };
  let cachedPrimary = THEME_CONFIG.spotify.primary;
  let cachedGlow = THEME_CONFIG.spotify.glow;
  let cachedGradient = null;

  function updateCachedGradient() {
    if (!canvas.height || !canvas.width) return;
    cachedGradient = ctx.createLinearGradient(0, canvas.height, 0, 0);
    cachedGradient.addColorStop(0, "rgba(255, 255, 255, 0.05)");
    cachedGradient.addColorStop(0.5, cachedPrimary);
    cachedGradient.addColorStop(1, "#ffffff");
  }

  function resizeCanvas() {
    canvas.width = canvas.parentElement.clientWidth * window.devicePixelRatio;
    canvas.height = canvas.parentElement.clientHeight * window.devicePixelRatio;
    updateCachedGradient();
  }
  window.addEventListener("resize", resizeCanvas);
  resizeCanvas();

  let phase = 0;
  function renderVisualizer() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const width = canvas.width;
    const height = canvas.height;
    const sensFactor = visualizerSensitivity / 80.0;

    phase += currentState.is_playing ? 0.07 : 0.02;

    if (visualizerMode === "wave") {
      ctx.beginPath();
      ctx.moveTo(0, height / 2);
      for (let x = 0; x < width; x += 5) {
        const freq = currentState.is_playing ? 0.015 : 0.008;
        const amp = (currentState.is_playing ? height * 0.35 : height * 0.12) * sensFactor;
        const y = height / 2 + Math.sin(x * freq + phase) * amp * Math.cos(x * 0.005 + phase * 0.5);
        ctx.lineTo(x, y);
      }
      // Zero-CPU dual-pass neon glow (avoids expensive CPU shadowBlur filters)
      ctx.strokeStyle = cachedGlow;
      ctx.lineWidth = 5 * window.devicePixelRatio;
      ctx.stroke();
      ctx.strokeStyle = cachedPrimary;
      ctx.lineWidth = 2 * window.devicePixelRatio;
      ctx.stroke();
    } else if (visualizerMode === "mirrored") {
      const barWidth = (width / numBars) * 0.65;
      const spacing = (width / numBars) * 0.35;
      const centerY = height / 2;

      ctx.fillStyle = cachedPrimary;
      for (let i = 0; i < numBars; i++) {
        const b = bars[i];
        if (currentState.is_playing) {
          const freqWave = Math.sin(phase + (i * 0.35)) * 0.5 + 0.5;
          const beatBounce = Math.sin(phase * 2.2 + (i % 4)) > 0.6 ? 1 : 0.2;
          b.targetHeight = (freqWave * 0.6 + beatBounce * 0.3) * (centerY * 0.85) * sensFactor;
        } else {
          b.targetHeight = (Math.sin(phase + i * 0.2) * 0.2 + 0.2) * (centerY * 0.3);
        }
        b.height += (b.targetHeight - b.height) * 0.25;

        const x = i * (barWidth + spacing) + spacing / 2;
        ctx.fillRect(x, centerY - b.height, barWidth, b.height * 2);
      }
    } else {
      const barWidth = (width / numBars) * 0.65;
      const spacing = (width / numBars) * 0.35;

      if (!cachedGradient) updateCachedGradient();
      ctx.fillStyle = cachedGradient;

      for (let i = 0; i < numBars; i++) {
        const b = bars[i];
        if (currentState.is_playing) {
          const freqWave = Math.sin(phase + (i * 0.35)) * 0.5 + 0.5;
          const beatBounce = Math.sin(phase * 2.2 + (i % 4)) > 0.6 ? 1 : 0.2;
          const noise = Math.random() * 0.25;
          b.targetHeight = (freqWave * 0.65 + beatBounce * 0.2 + noise) * (height * 0.85) * sensFactor;
        } else {
          b.targetHeight = (Math.sin(phase + i * 0.2) * 0.2 + 0.25) * (height * 0.25);
        }

        b.height += (b.targetHeight - b.height) * 0.25;

        if (b.height > b.peak) {
          b.peak = b.height;
          b.peakHold = 12;
        } else {
          if (b.peakHold > 0) {
            b.peakHold--;
          } else {
            b.peak -= 1.5 * window.devicePixelRatio;
            if (b.peak < b.height) b.peak = b.height;
          }
        }

        const x = i * (barWidth + spacing) + spacing / 2;
        const y = height - b.height;

        // Fast native GPU fillRect
        ctx.fillStyle = cachedGradient;
        ctx.fillRect(x, y, barWidth, b.height);

        // Peak line
        const peakY = height - b.peak - 3;
        ctx.fillStyle = "#ffffff";
        ctx.fillRect(x, Math.max(peakY, 0), barWidth, 2 * window.devicePixelRatio);
      }
    }

    requestAnimationFrame(renderVisualizer);
  }
  requestAnimationFrame(renderVisualizer);

  // 6. Settings Modal & Theme Accents (No AI!)
  function applyTheme(theme) {
    currentTheme = theme;
    document.body.setAttribute("data-theme", theme);
    if (THEME_CONFIG[theme]) {
      cachedPrimary = THEME_CONFIG[theme].primary;
      cachedGlow = THEME_CONFIG[theme].glow;
    }
    updateCachedGradient();
    themeSwatches.forEach(sw => {
      if (sw.getAttribute("data-theme") === theme) {
        sw.classList.add("active");
      } else {
        sw.classList.remove("active");
      }
    });
  }

  themeSwatches.forEach(sw => {
    sw.addEventListener("click", () => {
      applyTheme(sw.getAttribute("data-theme"));
    });
  });

  btnSettings.addEventListener("click", async () => {
    try {
      const res = await fetch("/api/config");
      if (res.ok) {
        const cfg = await res.json();
        cfgClockFormat.checked = cfg.clock_format_24h ?? true;
        cfgShowSeconds.checked = cfg.show_seconds ?? true;
        cfgVisualizer.value = cfg.visualizer_mode || "bars";
        cfgSensitivity.value = cfg.sensitivity || 80;
        if (cfg.theme_accent) applyTheme(cfg.theme_accent);
      }
    } catch (e) {}
    modalSettings.classList.remove("hidden");
  });

  btnCloseModal.addEventListener("click", () => modalSettings.classList.add("hidden"));
  modalSettings.addEventListener("click", (e) => {
    if (e.target === modalSettings) modalSettings.classList.add("hidden");
  });

  btnSaveConfig.addEventListener("click", async () => {
    clock24H = cfgClockFormat.checked;
    showSeconds = cfgShowSeconds.checked;
    visualizerMode = cfgVisualizer.value;
    visualizerSensitivity = parseInt(cfgSensitivity.value, 10) || 80;

    const payload = {
      clock_format_24h: clock24H,
      show_seconds: showSeconds,
      visualizer_mode: visualizerMode,
      theme_accent: currentTheme,
      sensitivity: visualizerSensitivity
    };

    try {
      await fetch("/api/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      modalSettings.classList.add("hidden");
    } catch (e) {
      alert("Fehler beim Speichern der Konfiguration.");
    }
  });
})();
