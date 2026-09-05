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
  const progressFillEl = document.getElementById("progressFill");
  const currentTimeEl = document.getElementById("currentTime");
  const totalTimeEl = document.getElementById("totalTime");

  const btnPrev = document.getElementById("btnPrev");
  const btnPlay = document.getElementById("btnPlay");
  const btnNext = document.getElementById("btnNext");
  const playIcon = document.getElementById("playIcon");
  const pauseIcon = document.getElementById("pauseIcon");
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
  const cfgClockFormat = document.getElementById("cfgClockFormat");
  const cfgAutoSkip = document.getElementById("cfgAutoSkip");
  const cfgAIKey = document.getElementById("cfgAIKey");
  const cfgAIUrl = document.getElementById("cfgAIUrl");

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
  let visualizerMode = "bars";

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
    secondsEl.textContent = s;

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
      statusDotEl.style.backgroundColor = "#1db954";
      statusDotEl.style.boxShadow = "0 0 10px #1db954";
    } else if (data.source === "spotify_player") {
      sourceLabelEl.textContent = "spotify_player CLI";
      statusDotEl.style.backgroundColor = "#1db954";
    } else {
      sourceLabelEl.textContent = "Demo / Standby";
      statusDotEl.style.backgroundColor = "#64748b";
      statusDotEl.style.boxShadow = "none";
    }
  }

  // Smooth local timeline ticker (60 FPS)
  function tickTimeline() {
    if (currentState.is_playing && currentState.duration_ms > 0) {
      const now = Date.now();
      const delta = now - lastSyncTime;
      const currentPos = Math.min(localPositionMs + delta, currentState.duration_ms);
      currentTimeEl.textContent = formatMs(currentPos);
      const pct = (currentPos / currentState.duration_ms) * 100;
      progressFillEl.style.width = `${pct}%`;
    } else {
      currentTimeEl.textContent = formatMs(localPositionMs);
      const pct = currentState.duration_ms > 0 ? (localPositionMs / currentState.duration_ms) * 100 : 0;
      progressFillEl.style.width = `${pct}%`;
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
      setTimeout(pollStatus, 200);
    } catch (e) {
      console.error(e);
    }
  }

  btnPlay.addEventListener("click", () => postAction("/api/playpause"));
  btnNext.addEventListener("click", () => postAction("/api/next"));
  btnPrev.addEventListener("click", () => postAction("/api/previous"));

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

  function resizeCanvas() {
    canvas.width = canvas.parentElement.clientWidth * window.devicePixelRatio;
    canvas.height = canvas.parentElement.clientHeight * window.devicePixelRatio;
  }
  window.addEventListener("resize", resizeCanvas);
  resizeCanvas();

  let phase = 0;
  function renderVisualizer() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const width = canvas.width;
    const height = canvas.height;
    const barWidth = (width / numBars) * 0.65;
    const spacing = (width / numBars) * 0.35;

    phase += currentState.is_playing ? 0.08 : 0.02;

    for (let i = 0; i < numBars; i++) {
      const b = bars[i];
      if (currentState.is_playing) {
        // Dynamic simulated FFT frequency bands
        const freqWave = Math.sin(phase + (i * 0.35)) * 0.5 + 0.5;
        const beatBounce = Math.sin(phase * 2.2 + (i % 4)) > 0.6 ? 1 : 0.2;
        const noise = Math.random() * 0.25;
        b.targetHeight = (freqWave * 0.65 + beatBounce * 0.2 + noise) * (height * 0.85);
      } else {
        // Subtle resting sine wave
        b.targetHeight = (Math.sin(phase + i * 0.2) * 0.2 + 0.25) * (height * 0.25);
      }

      // Smooth interpolation with peak drop
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

      // Bar gradient
      const grad = ctx.createLinearGradient(0, height, 0, y);
      grad.addColorStop(0, "rgba(29, 185, 84, 0.2)");
      grad.addColorStop(0.5, "rgba(29, 185, 84, 0.85)");
      grad.addColorStop(1, "rgba(56, 189, 248, 0.95)");

      ctx.fillStyle = grad;
      ctx.beginPath();
      ctx.roundRect(x, y, barWidth, b.height, [4, 4, 0, 0]);
      ctx.fill();

      // Peak cap
      const peakY = height - b.peak - 3;
      ctx.fillStyle = "#38bdf8";
      ctx.fillRect(x, Math.max(peakY, 0), barWidth, 2 * window.devicePixelRatio);
    }

    requestAnimationFrame(renderVisualizer);
  }
  requestAnimationFrame(renderVisualizer);

  // 6. Settings Modal
  btnSettings.addEventListener("click", async () => {
    try {
      const res = await fetch("/api/config");
      if (res.ok) {
        const cfg = await res.json();
        cfgClockFormat.checked = cfg.clock_format_24h ?? true;
        cfgAutoSkip.checked = cfg.auto_skip_depri ?? false;
        cfgVisualizer.value = cfg.visualizer_mode || "bars";
        cfgAIUrl.value = cfg.ai_base_url || "http://localhost:9001";
        if (cfg.has_ai_key) {
          cfgAIKey.placeholder = "•••••••••••• (gespeichert)";
        }
      }
    } catch (e) {}
    modalSettings.classList.remove("hidden");
  });

  btnCloseModal.addEventListener("click", () => modalSettings.classList.add("hidden"));
  modalSettings.addEventListener("click", (e) => {
    if (e.target === modalSettings) modalSettings.classList.add("hidden");
  });

  btnSaveConfig.addEventListener("click", async () => {
    const payload = {
      clock_format_24h: cfgClockFormat.checked,
      auto_skip_depri: cfgAutoSkip.checked,
      visualizer_mode: cfgVisualizer.value,
      ai_base_url: cfgAIUrl.value
    };
    if (cfgAIKey.value.trim() !== "") {
      payload.ai_api_key = cfgAIKey.value.trim();
    }
    clock24H = cfgClockFormat.checked;
    visualizerMode = cfgVisualizer.value;

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
