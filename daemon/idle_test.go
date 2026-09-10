package daemon

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestIdleMonitorLifecycle(t *testing.T) {
	var triggered int32
	mon := NewMonitor(1, func() {
		atomic.AddInt32(&triggered, 1)
	})

	mon.Start()
	time.Sleep(100 * time.Millisecond)
	mon.Stop()

	// Should stop cleanly without hanging or panic
}
