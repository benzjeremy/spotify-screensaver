//go:build windows

package audio

// WindowsCaptureEngine provides audio capture for Windows via fallback synthesizer / WASAPI stub.
type WindowsCaptureEngine struct {
	fallback *FallbackSynthesizer
}

// NewPlatformCaptureEngine creates a Windows audio engine.
func NewPlatformCaptureEngine() Engine {
	return &WindowsCaptureEngine{
		fallback: NewFallbackSynthesizer(),
	}
}

func (w *WindowsCaptureEngine) Start() error {
	return w.fallback.Start()
}

func (w *WindowsCaptureEngine) Stop() {
	w.fallback.Stop()
}

func (w *WindowsCaptureEngine) GetBands() [NumBands]byte {
	return w.fallback.GetBands()
}
