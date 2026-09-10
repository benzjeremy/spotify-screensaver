package daemon

import (
	"log"
	"sync"
	"time"
)

// Monitor tracks user idle time and fires a callback after inactivity.
type Monitor struct {
	mu         sync.Mutex
	timeoutSec int
	onIdle     func()
	stopChan   chan struct{}
	running    bool
}

// NewMonitor creates an idle monitor.
func NewMonitor(timeoutSec int, onIdle func()) *Monitor {
	return &Monitor{
		timeoutSec: timeoutSec,
		onIdle:     onIdle,
		stopChan:   make(chan struct{}),
	}
}

// Start begins monitoring user idle time in a background goroutine.
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running || m.timeoutSec <= 0 {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	log.Printf("[Idle-Daemon] Inaktivitätsüberwachung aktiv (%d Sekunden Timeout)\n", m.timeoutSec)

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				return
			case <-ticker.C:
				idleDuration := getSystemIdleDuration()
				if idleDuration >= time.Duration(m.timeoutSec)*time.Second {
					if m.onIdle != nil {
						m.onIdle()
					}
				}
			}
		}
	}()
}

// Stop terminates the idle monitor.
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		close(m.stopChan)
		m.running = false
	}
}
