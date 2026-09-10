//go:build linux

package daemon

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func getSystemIdleDuration() time.Duration {
	// 1. Try xprintidle (X11)
	out, err := exec.Command("xprintidle").Output()
	if err == nil {
		if ms, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
	}

	// 2. Try GNOME Mutter IdleMonitor via gdbus
	// gdbus call --session --dest org.gnome.Mutter.IdleMonitor --object-path /org/gnome/Mutter/IdleMonitor/Core --method org.gnome.Mutter.IdleMonitor.GetIdletime
	out, err = exec.Command("gdbus", "call", "--session", "--dest", "org.gnome.Mutter.IdleMonitor",
		"--object-path", "/org/gnome/Mutter/IdleMonitor/Core",
		"--method", "org.gnome.Mutter.IdleMonitor.GetIdletime").Output()
	if err == nil {
		str := strings.TrimSpace(string(out))
		// Format: (uint64 12345,)
		if strings.HasPrefix(str, "(uint64 ") {
			str = strings.TrimPrefix(str, "(uint64 ")
			if idx := strings.Index(str, ","); idx != -1 {
				str = str[:idx]
			}
			str = strings.TrimSuffix(str, ")")
			if ms, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64); err == nil {
				return time.Duration(ms) * time.Millisecond
			}
		}
	}

	return 0
}
