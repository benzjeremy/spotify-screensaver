//go:build !linux && !windows

package audio

type OtherCaptureEngine struct {
	fallback *FallbackSynthesizer
}

func NewPlatformCaptureEngine() Engine {
	return &OtherCaptureEngine{
		fallback: NewFallbackSynthesizer(),
	}
}

func (o *OtherCaptureEngine) Start() error {
	return o.fallback.Start()
}

func (o *OtherCaptureEngine) Stop() {
	o.fallback.Stop()
}

func (o *OtherCaptureEngine) GetBands() [NumBands]byte {
	return o.fallback.GetBands()
}
