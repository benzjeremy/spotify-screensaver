//go:build !linux && !windows

package daemon

import "time"

func getSystemIdleDuration() time.Duration {
	return 0
}
