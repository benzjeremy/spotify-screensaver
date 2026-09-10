package audio

import (
	"math"
	"sync"
	"time"
)

// Engine defines the common interface for audio capture engines.
type Engine interface {
	Start() error
	Stop()
	GetBands() [NumBands]byte
}

// FallbackSynthesizer generates responsive simulated audio spectrum when no microphone/sink is active.
type FallbackSynthesizer struct {
	mu        sync.RWMutex
	analyzer  *FFTAnalyzer
	isPlaying bool
	phase     float64
	bands     [NumBands]byte
	stopChan  chan struct{}
}

// NewFallbackSynthesizer creates a synthetic spectrum generator.
func NewFallbackSynthesizer() *FallbackSynthesizer {
	return &FallbackSynthesizer{
		analyzer: NewFFTAnalyzer(),
		stopChan: make(chan struct{}),
	}
}

// SetPlaying informs the synthesizer whether Spotify is currently active.
func (fs *FallbackSynthesizer) SetPlaying(playing bool) {
	fs.mu.Lock()
	fs.isPlaying = playing
	fs.mu.Unlock()
}

func (fs *FallbackSynthesizer) Start() error {
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()

		samples := make([]float64, FFTSize)
		for {
			select {
			case <-fs.stopChan:
				return
			case <-ticker.C:
				fs.mu.Lock()
				playing := fs.isPlaying
				fs.phase += 0.08
				p := fs.phase
				fs.mu.Unlock()

				// Generate synthetic musical frequencies (Bass kick + Mids harmonic + High noise)
				for i := 0; i < FFTSize; i++ {
					t := float64(i) / 44100.0
					if playing {
						bass := math.Sin(2*math.Pi*60.0*t + p*4.0)
						snare := math.Sin(2*math.Pi*250.0*t+p*2.0) * math.Cos(p*1.5)
						hihat := math.Sin(2*math.Pi*4000.0*t) * (math.Sin(p*8.0)*0.5 + 0.5)
						samples[i] = (bass*0.5 + snare*0.35 + hihat*0.15) * 0.7
					} else {
						// Ambient gentle idling wave
						samples[i] = math.Sin(2*math.Pi*120.0*t+p) * 0.12
					}
				}

				bands := fs.analyzer.ComputeBands(samples)
				fs.mu.Lock()
				fs.bands = bands
				fs.mu.Unlock()
			}
		}
	}()
	return nil
}

func (fs *FallbackSynthesizer) Stop() {
	select {
	case <-fs.stopChan:
	default:
		close(fs.stopChan)
	}
}

func (fs *FallbackSynthesizer) GetBands() [NumBands]byte {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.bands
}
