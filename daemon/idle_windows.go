//go:build windows

package daemon

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetLastInputInfo = user32.NewProc("GetLastInputInfo")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount     = kernel32.NewProc("GetTickCount")
)

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

func getSystemIdleDuration() time.Duration {
	var lii lastInputInfo
	lii.cbSize = uint32(unsafe.Sizeof(lii))

	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0
	}

	tickCount, _, _ := procGetTickCount.Call()
	elapsedMs := uint32(tickCount) - lii.dwTime
	return time.Duration(elapsedMs) * time.Millisecond
}
