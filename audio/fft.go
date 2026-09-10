package audio

import (
	"math"
	"math/cmplx"
)

// FFTSize is the number of audio samples per FFT frame (must be power of 2).
const FFTSize = 1024

// NumBands is the standard number of frequency spectrum bands.
const NumBands = 64

// FFTAnalyzer performs Fast Fourier Transforms and aggregates 64 frequency bands.
type FFTAnalyzer struct {
	hannWindow []float64
	bands      [NumBands]float64
	decayRate  float64
}

// NewFFTAnalyzer initializes an analyzer with pre-computed windowing.
func NewFFTAnalyzer() *FFTAnalyzer {
	window := make([]float64, FFTSize)
	for i := 0; i < FFTSize; i++ {
		window[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(FFTSize-1)))
	}
	return &FFTAnalyzer{
		hannWindow: window,
		decayRate:  0.82,
	}
}

// ComputeBands takes raw float64 audio samples (normalized -1.0 to 1.0) and returns 64 normalized byte values (0-255).
func (fa *FFTAnalyzer) ComputeBands(samples []float64) [NumBands]byte {
	in := make([]complex128, FFTSize)
	n := len(samples)
	if n > FFTSize {
		n = FFTSize
	}

	for i := 0; i < n; i++ {
		in[i] = complex(samples[i]*fa.hannWindow[i], 0)
	}

	// Compute Radix-2 FFT
	fftRadix2(in)

	// Half of the spectrum represents frequencies from 0 to Nyquist (22050 Hz for 44.1kHz)
	halfN := FFTSize / 2
	magnitudes := make([]float64, halfN)
	for i := 0; i < halfN; i++ {
		magnitudes[i] = cmplx.Abs(in[i]) / float64(FFTSize)
	}

	// Map to 64 logarithmically spaced frequency bands
	var rawBands [NumBands]float64
	for b := 0; b < NumBands; b++ {
		// Logarithmic distribution from bin 1 to halfN-1
		lowP := math.Pow(float64(b)/float64(NumBands), 2.2)
		highP := math.Pow(float64(b+1)/float64(NumBands), 2.2)

		lowBin := int(lowP * float64(halfN-2))
		highBin := int(highP * float64(halfN-1))
		if highBin <= lowBin {
			highBin = lowBin + 1
		}
		if highBin > halfN {
			highBin = halfN
		}

		sum := 0.0
		count := 0
		for bin := lowBin; bin < highBin; bin++ {
			sum += magnitudes[bin]
			count++
		}

		val := 0.0
		if count > 0 {
			val = sum / float64(count)
		}

		// Apply perceptual equalizer curve (boost higher frequencies)
		freqBoost := 1.0 + float64(b)*0.08
		val *= freqBoost * 14.0

		// Non-linear amplitude scaling (sqrt/log-like)
		if val > 0 {
			val = math.Sqrt(val)
		}
		if val > 1.0 {
			val = 1.0
		}
		rawBands[b] = val
	}

	// Temporal smoothing (Attack & Decay)
	var out [NumBands]byte
	for b := 0; b < NumBands; b++ {
		target := rawBands[b]
		if target > fa.bands[b] {
			// Fast attack
			fa.bands[b] = fa.bands[b]*0.35 + target*0.65
		} else {
			// Smooth exponential decay
			fa.bands[b] = fa.bands[b] * fa.decayRate
		}

		bVal := int(fa.bands[b] * 255.0)
		if bVal < 0 {
			bVal = 0
		} else if bVal > 255 {
			bVal = 255
		}
		out[b] = byte(bVal)
	}

	return out
}

// In-place Radix-2 Cooley-Tukey FFT algorithm
func fftRadix2(a []complex128) {
	n := len(a)
	if n <= 1 {
		return
	}

	// Bit-reversal permutation
	j := 0
	for i := 0; i < n; i++ {
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
		bit := n >> 1
		for (j & bit) != 0 {
			j ^= bit
			bit >>= 1
		}
		j ^= bit
	}

	// Cooley-Tukey computation
	for lenStep := 2; lenStep <= n; lenStep <<= 1 {
		angle := -2.0 * math.Pi / float64(lenStep)
		wlen := cmplx.Rect(1.0, angle)
		for i := 0; i < n; i += lenStep {
			w := complex(1.0, 0)
			for k := 0; k < lenStep/2; k++ {
				u := a[i+k]
				v := a[i+k+lenStep/2] * w
				a[i+k] = u + v
				a[i+k+lenStep/2] = u - v
				w *= wlen
			}
		}
	}
}
