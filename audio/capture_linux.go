//go:build linux

package audio

import (
	"encoding/binary"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// LinuxCaptureEngine captures live audio from default output monitor using parec or pw-record.
type LinuxCaptureEngine struct {
	mu        sync.RWMutex
	analyzer  *FFTAnalyzer
	cmd       *exec.Cmd
	bands     [NumBands]byte
	stopChan  chan struct{}
	fallback  *FallbackSynthesizer
	isLive    bool
	isPlaying bool
}

// NewPlatformCaptureEngine creates a Linux capture engine.
func NewPlatformCaptureEngine() Engine {
	return &LinuxCaptureEngine{
		analyzer: NewFFTAnalyzer(),
		stopChan: make(chan struct{}),
		fallback: NewFallbackSynthesizer(),
	}
}

// SetPlaying sets playing status for fallback synthesizer.
func (l *LinuxCaptureEngine) SetPlaying(playing bool) {
	l.mu.Lock()
	l.isPlaying = playing
	l.mu.Unlock()
	l.fallback.SetPlaying(playing)
}

func (l *LinuxCaptureEngine) Start() error {
	l.fallback.Start()

	go l.runCapture()
	return nil
}

func (l *LinuxCaptureEngine) runCapture() {
	// Find monitor source
	sinkOut, _ := exec.Command("pactl", "get-default-sink").Output()
	sinkName := strings.TrimSpace(string(sinkOut))
	monitorSource := "@DEFAULT_SINK@.monitor"
	if sinkName != "" {
		monitorSource = sinkName + ".monitor"
	}

	cmd := exec.Command("parec", "-d", monitorSource, "--format=s16le", "--rate=44100", "--channels=2")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Audio] parec konnte nicht gestartet werden: %v (nutze reaktiven Fallback)\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[Audio] Fehler beim Starten von parec: %v (nutze reaktiven Fallback)\n", err)
		return
	}

	l.cmd = cmd
	log.Printf("[Audio] Live Audio-Monitor aktiv via %s\n", monitorSource)

	// Buffer: 1024 samples * 2 channels * 2 bytes = 4096 bytes per frame
	bufSize := FFTSize * 2 * 2
	rawBuf := make([]byte, bufSize)
	samples := make([]float64, FFTSize)

	// Watchdog for audio silence / disconnected state
	go func() {
		<-l.stopChan
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	for {
		select {
		case <-l.stopChan:
			return
		default:
		}

		_, err := io.ReadFull(stdout, rawBuf)
		if err != nil {
			l.mu.Lock()
			l.isLive = false
			l.mu.Unlock()
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// Convert stereo 16-bit PCM to mono float64 (-1.0 to 1.0)
		hasSignal := false
		for i := 0; i < FFTSize; i++ {
			left := int16(binary.LittleEndian.Uint16(rawBuf[i*4 : i*4+2]))
			right := int16(binary.LittleEndian.Uint16(rawBuf[i*4+2 : i*4+4]))
			avg := (float64(left) + float64(right)) / (2.0 * 32768.0)
			if avg > 0.005 || avg < -0.005 {
				hasSignal = true
			}
			samples[i] = avg
		}

		if hasSignal {
			bands := l.analyzer.ComputeBands(samples)
			l.mu.Lock()
			l.bands = bands
			l.isLive = true
			l.mu.Unlock()
		} else {
			// If silence on live stream, check if Spotify is playing, else fallback
			l.mu.Lock()
			l.isLive = false
			l.mu.Unlock()
		}
	}
}

func (l *LinuxCaptureEngine) Stop() {
	select {
	case <-l.stopChan:
	default:
		close(l.stopChan)
	}
	if l.cmd != nil && l.cmd.Process != nil {
		l.cmd.Process.Kill()
	}
	l.fallback.Stop()
}

func (l *LinuxCaptureEngine) GetBands() [NumBands]byte {
	l.mu.RLock()
	live := l.isLive
	bands := l.bands
	l.mu.RUnlock()

	if live {
		return bands
	}
	return l.fallback.GetBands()
}
