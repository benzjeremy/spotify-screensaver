//go:build (!linux && !windows) || (linux && !cgo)

package main

import (
	"log"
	"os/exec"
)

func LaunchGUI(title, url string, width, height int, fullscreen bool) {
	log.Printf("[GUI] Öffne Standardbrowser (%s)...\n", url)
	_ = exec.Command("xdg-open", url).Start()
}
