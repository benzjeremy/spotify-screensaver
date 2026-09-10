package audio

import (
	"math"
	"testing"
	"time"
)

func TestFFTAnalyzerComputeBands(t *testing.T) {
	analyzer := NewFFTAnalyzer()

	// Generate a 440 Hz test tone at 44100 Hz sample rate
	samples := make([]float64, FFTSize)
	for i := 0; i < FFTSize; i++ {
		tVal := float64(i) / 44100.0
		samples[i] = math.Sin(2 * math.Pi * 440.0 * tVal)
	}

	bands := analyzer.ComputeBands(samples)

	// Check that we received NumBands bands
	if len(bands) != NumBands {
		t.Fatalf("Expected %d bands, got %d", NumBands, len(bands))
	}

	// 440 Hz should produce non-zero energy in lower-mid bands
	hasEnergy := false
	for _, b := range bands {
		if b > 0 {
			hasEnergy = true
			break
		}
	}

	if !hasEnergy {
		t.Fatalf("Expected non-zero energy from 440 Hz tone")
	}
}

func TestFallbackSynthesizer(t *testing.T) {
	synth := NewFallbackSynthesizer()
	synth.SetPlaying(true)
	if err := synth.Start(); err != nil {
		t.Fatalf("Failed to start synthesizer: %v", err)
	}
	defer synth.Stop()

	// Wait for ticker to generate at least one frame
	time.Sleep(50 * time.Millisecond)

	bands := synth.GetBands()
	hasEnergy := false
	for _, b := range bands {
		if b > 0 {
			hasEnergy = true
			break
		}
	}

	if !hasEnergy {
		t.Fatalf("Expected synthesizer to produce frequency bands")
	}
}
